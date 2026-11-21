package agent

import "github.com/madnikulin50/lowcode/agents/sdk"

func field(key, widget, label string, extra ...func(*sdk.Field)) sdk.Field {
	out := sdk.Field{Key: key, Widget: widget, Label: label}
	for _, fn := range extra {
		fn(&out)
	}
	return out
}

func req(f *sdk.Field)  { f.Required = true }
func tmpl(f *sdk.Field) { f.Template = true }

func Components() []sdk.Component {
	return []sdk.Component{
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/run", Label: "Backup: run", Description: "Start a backup job from a source or policy",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "backup",
			ConfigFields: []sdk.Field{
				field("sourceID", "string", "Source ID", tmpl),
				field("policyID", "string", "Policy ID", tmpl),
				field("destinationID", "string", "Destination ID (ad-hoc, пусто = из политики)", tmpl),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/dest-test", Label: "Backup: проверить назначение", Description: "Проверить endpoint/учётки/бакет назначения и записать результат в запись",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: false,
			Service: "backup", Operation: "dest-test",
			ConfigFields: []sdk.Field{
				field("destinationID", "string", "Destination ID", req, tmpl),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/restore", Label: "Backup: restore", Description: "Restore a snapshot",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "restore",
			ConfigFields: []sdk.Field{
				field("snapshotID", "string", "Snapshot ID", req, tmpl),
				field("destType", "string", "Destination type", tmpl),
				field("destPath", "string", "Destination path", tmpl),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/prune", Label: "Backup: prune", Description: "Apply retention and delete expired snapshots",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "prune",
			ConfigFields: []sdk.Field{
				field("policyID", "string", "Policy ID", tmpl),
				field("sourceID", "string", "Source ID", tmpl),
				field("retentionDays", "number", "Retention days"),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/due", Label: "Backup: run due", Description: "Run policies whose cron matches now",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: false,
			Service: "backup", Operation: "due",
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/vault", Label: "Backup: в сейф", Description: "Сканировать снапшот и скопировать в WORM-сейф",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "vault",
			ConfigFields: []sdk.Field{
				field("snapshotID", "string", "Snapshot ID", req, tmpl),
				field("retentionDays", "number", "Retention дней (перекрывает политику)"),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/scan", Label: "Backup: сканировать", Description: "Сканировать снапшот без копии в сейф",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "scan",
			ConfigFields: []sdk.Field{
				field("snapshotID", "string", "Snapshot ID", req, tmpl),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/clean-point", Label: "Backup: чистая точка", Description: "Последняя чистая точка восстановления (после атаки шифровальщика)",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: false,
			Service: "backup", Operation: "clean-point",
			ConfigFields: []sdk.Field{
				field("sourceID", "string", "Source ID (optional)", tmpl),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/verify", Label: "Backup: верификация", Description: "Поднять снапшот в изолированной песочнице и проверить восстановление",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "verify",
			ConfigFields: []sdk.Field{
				field("snapshotID", "string", "Snapshot ID", tmpl),
				field("policyID", "string", "Policy ID (верифицирует последний снапшот)", tmpl),
				field("verifyLevel", "string", "Уровень: l1 (целостность) | l2 (+ docker)", tmpl),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/verify-due", Label: "Backup: верификация due", Description: "Снапшоты, у которых по расписанию пора верификация",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: false,
			Service: "backup", Operation: "verify-due",
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/browse", Label: "Backup: содержимое", Description: "Показать, что лежит внутри снапшота (файлы, письма, объекты)",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: false,
			Service: "backup", Operation: "browse",
			ConfigFields: []sdk.Field{
				field("snapshotID", "string", "Snapshot ID", req, tmpl),
				field("path", "string", "Префикс пути (optional)", tmpl),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/extract", Label: "Backup: достать файл", Description: "Извлечь один файл/папку из снапшота (без полного восстановления)",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "extract",
			ConfigFields: []sdk.Field{
				field("snapshotID", "string", "Snapshot ID", req, tmpl),
				field("path", "string", "Путь внутри снапшота", req, tmpl),
				field("extractTo", "string", "Куда писать (пусто = вернуть текст)", tmpl),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/audit", Label: "Backup: аудит данных", Description: "Найти данные, которые не покрыты ни одним источником бэкапа",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "audit",
			ConfigFields: []sdk.Field{
				field("scopeID", "string", "Область аудита (пусто = все включённые)", tmpl),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/compliance", Label: "Backup: соответствие", Description: "Сводка фактического состояния для плана резервирования (Приказ №117)",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: false,
			Service: "backup", Operation: "compliance",
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/archive-manifest", Label: "Backup: манифест архива", Description: "Подписанный перечень файлов с SHA-256 (юридическая значимость архива)",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "manifest",
			ConfigFields: []sdk.Field{
				field("snapshotID", "string", "Snapshot ID", req, tmpl),
			},
		}},
		sdk.Desc{D: sdk.Descriptor{
			Type: "backup/destroy", Label: "Backup: уничтожить снапшот", Description: "Уничтожение по истечении срока хранения (основание обязательно; legal hold блокирует)",
			Category: sdk.CategoryAction, Execution: sdk.ExecRemote, Async: true,
			Service: "backup", Operation: "destroy",
			ConfigFields: []sdk.Field{
				field("snapshotID", "string", "Snapshot ID", req, tmpl),
				field("basis", "string", "Основание (срок хранения / приказ)", req, tmpl),
			},
		}},
	}
}
