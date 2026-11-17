package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// K8s-бэкап: снапшот состояния кластера (манифесты всех namespaced-объектов,
// при include_pv — и PV) в JSON-файлы, упакованные тем же tar.gz-пайплайном.
// Данные томов (PVC content) сюда не входят — это фаза 2 (CSI/restic с нод).
// Восстановление: destType=cluster применяет манифесты к целевому кластеру
// (kubeconfig из credentials записи restore/источника) в порядке зависимостей.

// k8sClientBundle держит dynamic-клиент и discovery для одного kubeconfig.
type k8sClientBundle struct {
	dyn   dynamic.Interface
	disco discovery.DiscoveryInterface
}

func k8sClients(cred *Credential) (*k8sClientBundle, error) {
	raw := strings.TrimSpace(cred.Extra)
	if raw == "" {
		return nil, fmt.Errorf("kubernetes: empty kubeconfig in credentials.extra")
	}
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && strings.Contains(string(decoded), "apiVersion") {
		raw = string(decoded)
	}
	restCfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(raw))
	if err != nil {
		return nil, fmt.Errorf("kubernetes: kubeconfig: %w", err)
	}
	if restCfg.Timeout == 0 {
		restCfg.Timeout = 60 * time.Second
	}
	dyn, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}
	disco, err := discovery.NewDiscoveryClientForConfig(restCfg)
	if err != nil {
		return nil, err
	}
	return &k8sClientBundle{dyn: dyn, disco: disco}, nil
}

// k8sDump пишет манифесты в destDir в виде <namespace>/<resource>.<name>.json.
func k8sDump(ctx context.Context, src *Source, cred *Credential, destDir string) (int, error) {
	cl, err := k8sClients(cred)
	if err != nil {
		return 0, err
	}
	nsFilter := map[string]bool{}
	for _, n := range strings.Split(src.Namespaces, ",") {
		n = strings.TrimSpace(n)
		if n != "" {
			nsFilter[n] = true
		}
	}

	resources, err := cl.disco.ServerPreferredNamespacedResources()
	if err != nil {
		// Часть групп может быть недоступна — продолжаем с тем, что есть.
		if len(resources) == 0 {
			return 0, fmt.Errorf("discovery: %w", err)
		}
	}
	var gvs []schema.GroupVersionResource
	for _, list := range resources {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}
		for _, r := range list.APIResources {
			if strings.Contains(r.Name, "/") { // подресурсы пропускаем
				continue
			}
			gvs = append(gvs, gv.WithResource(r.Name))
		}
	}
	sort.Slice(gvs, func(i, j int) bool { return gvs[i].String() < gvs[j].String() })

	written := 0
	for _, gvr := range gvs {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		list, err := cl.dyn.Resource(gvr).List(ctx, metav1.ListOptions{Limit: 500})
		if err != nil {
			continue // нет прав на ресурс — не роняем выгрузку
		}
		for i := range list.Items {
			obj := list.Items[i]
			ns := obj.GetNamespace()
			if len(nsFilter) > 0 && !nsFilter[ns] {
				continue
			}
			dir := path.Join(destDir, ns)
			if ns == "" {
				dir = path.Join(destDir, "_cluster")
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return written, err
			}
			file := path.Join(dir, fmt.Sprintf("%s.%s.json", gvr.Resource, obj.GetName()))
			data, err := json.Marshal(obj.Object)
			if err != nil {
				continue
			}
			if err := os.WriteFile(file, data, 0o600); err != nil {
				return written, err
			}
			written++
		}
	}

	if src.IncludePV {
		n, err := k8sDumpClusterScoped(ctx, cl, destDir)
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

// k8sDumpClusterScoped выгружает PV и StorageClass.
func k8sDumpClusterScoped(ctx context.Context, cl *k8sClientBundle, destDir string) (int, error) {
	written := 0
	for _, res := range []string{"persistentvolumes", "storageclasses"} {
		gvr := schema.GroupVersionResource{Resource: res}
		if res == "storageclasses" {
			gvr = schema.GroupVersionResource{Group: "storage.k8s.io", Version: "v1", Resource: res}
		}
		list, err := cl.dyn.Resource(gvr).List(ctx, metav1.ListOptions{Limit: 500})
		if err != nil {
			continue
		}
		dir := path.Join(destDir, "_cluster")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return written, err
		}
		for i := range list.Items {
			obj := list.Items[i]
			file := path.Join(dir, fmt.Sprintf("%s.%s.json", res, obj.GetName()))
			data, _ := json.Marshal(obj.Object)
			if err := os.WriteFile(file, data, 0o600); err != nil {
				return written, err
			}
			written++
		}
	}
	return written, nil
}

// k8sApply восстанавливает манифесты из распакованной директории в кластер.
// Порядок: namespaces -> _cluster -> остальное по алфавиту.
func k8sApply(ctx context.Context, cred *Credential, srcDir string) (int, error) {
	cl, err := k8sClients(cred)
	if err != nil {
		return 0, err
	}
	type item struct {
		gvr schema.GroupVersionResource
		obj *unstructured.Unstructured
	}
	var first, cluster, rest []item
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return 0, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		nsName := e.Name()
		files, _ := os.ReadDir(path.Join(srcDir, nsName))
		for _, f := range files {
			if !strings.HasSuffix(f.Name(), ".json") {
				continue
			}
			raw, err := os.ReadFile(path.Join(srcDir, nsName, f.Name()))
			if err != nil {
				continue
			}
			var obj unstructured.Unstructured
			if err := json.Unmarshal(raw, &obj); err != nil {
				continue
			}
			gvr, err := k8sGVRFor(&obj)
			if err != nil {
				continue
			}
			it := item{gvr: gvr, obj: &obj}
			switch {
			case gvr.Resource == "namespaces":
				first = append(first, it)
			case nsName == "_cluster":
				cluster = append(cluster, it)
			default:
				rest = append(rest, it)
			}
		}
	}
	applied := 0
	for _, group := range [][]item{first, cluster, rest} {
		for _, it := range group {
			if err := ctx.Err(); err != nil {
				return applied, err
			}
			var err error
			if it.gvr.Resource == "namespaces" {
				_, err = cl.dyn.Resource(it.gvr).Create(ctx, it.obj, metav1.CreateOptions{})
			} else {
				_, err = cl.dyn.Resource(it.gvr).Namespace(it.obj.GetNamespace()).
					Create(ctx, it.obj, metav1.CreateOptions{})
			}
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				return applied, fmt.Errorf("apply %s/%s: %w", it.gvr.Resource, it.obj.GetName(), err)
			}
			applied++
		}
	}
	return applied, nil
}

// k8sGVRFor восстанавливает GroupVersionResource из apiVersion/kind через
// discovery; для типовых ресурсов — через безопасный маппинг без discovery.
func k8sGVRFor(obj *unstructured.Unstructured) (schema.GroupVersionResource, error) {
	apiVersion := obj.GetAPIVersion()
	kind := obj.GetKind()
	if apiVersion == "" || kind == "" {
		return schema.GroupVersionResource{}, fmt.Errorf("no apiVersion/kind")
	}
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	// plural mapping для частых ресурсов
	plural := map[string]string{
		"Namespace": "namespaces", "Pod": "pods", "Deployment": "deployments",
		"StatefulSet": "statefulsets", "DaemonSet": "daemonsets", "Service": "services",
		"ConfigMap": "configmaps", "Secret": "secrets", "Ingress": "ingresses",
		"PersistentVolumeClaim": "persistentvolumeclaims", "PersistentVolume": "persistentvolumes",
		"ServiceAccount": "serviceaccounts", "Role": "roles", "RoleBinding": "rolebindings",
		"ClusterRole": "clusterroles", "ClusterRoleBinding": "clusterrolebindings",
		"Job": "jobs", "CronJob": "cronjobs",
	}[kind]
	if plural == "" {
		plural = strings.ToLower(kind) + "s"
	}
	return gv.WithResource(plural), nil
}

// k8sToMinio снимает манифесты кластера во временную директорию и упаковывает
// её тем же tar.gz-пайплайном, что и fs-источники.
func (a *Agent) k8sToMinio(ctx context.Context, st *JobStatus, store Storage, src *Source, cred *Credential, key string) (*BackupResult, error) {
	tmp, err := os.MkdirTemp("", "backup-k8s-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	st.Message = "dumping cluster manifests"
	written, err := k8sDump(ctx, src, cred, tmp)
	if err != nil {
		return nil, err
	}
	if written == 0 {
		return nil, fmt.Errorf("kubernetes: 0 objects dumped (check namespaces filter / RBAC)")
	}
	st.Files = written
	walker, err := NewLocalFS(tmp)
	if err != nil {
		return nil, err
	}
	res, err := a.packToMinio(ctx, st, store, walker, key)
	if err == nil {
		res.Engine = "k8s"
	}
	return res, err
}

// restoreToCluster скачивает снапшот, распаковывает и применяет манифесты.
func (a *Agent) restoreToCluster(ctx context.Context, src *Source, snap *Snapshot, cred *Credential, store Storage) error {
	rc, err := store.GetBucket(ctx, snap.S3Bucket, snap.S3Key)
	if err != nil {
		return err
	}
	defer rc.Close()
	tmp, err := os.MkdirTemp("", "backup-k8s-restore-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if _, err := UnpackTarGz(ctx, rc, tmp); err != nil {
		return err
	}
	applied, err := k8sApply(ctx, cred, tmp)
	if err != nil {
		return err
	}
	log.Printf("k8s restore snapshot %d: %d objects applied", snap.ID, applied)
	return nil
}
