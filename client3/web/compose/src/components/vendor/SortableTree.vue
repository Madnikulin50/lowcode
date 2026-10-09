<template>
  <div class="sortable-tree" :draggable="draggable && !!parentData" @dragstart.stop="dragStart" @dragover.stop.prevent @dragenter.stop.prevent="dragEnter" @dragleave.stop="dragLeave" @drop.stop.prevent="drop" @dragend.stop.prevent="dragEnd">
    <div class="content">
      <slot name="default" :item="data">
        <span>{{ data[attr] }}</span>
      </slot>
    </div>
    <ul v-if="showChildren">
      <li v-for="(child, index) in computedChildren" :key="child._key" :class="liClass(child)">
        <div v-if="child._replaceLi_" class="sortable-tree blank-placeholder" />
        <div v-else class="sortable-tree" :draggable="draggable && !!parentData" @dragstart.stop="dragStart" @dragover.stop.prevent @dragenter.stop.prevent="dragEnter" @dragleave.stop="dragLeave" @drop.stop.prevent="drop" @dragend.stop.prevent="dragEnd">
          <div class="content">
            <slot name="default" :item="child">
              <span>{{ child[attr] }}</span>
            </slot>
          </div>
          <ul v-if="hasChildren(child)">
            <li v-for="(sub, subIndex) in childChildren(child)" :key="sub._key" :class="liClass(sub)">
              <div v-if="sub._replaceLi_" class="sortable-tree blank-placeholder" />
              <div v-else class="sortable-tree" :draggable="draggable && !!parentData" @dragstart.stop="dragStart" @dragover.stop.prevent @dragenter.stop.prevent="dragEnter" @dragleave.stop="dragLeave" @drop.stop.prevent="drop" @dragend.stop.prevent="dragEnd">
                <div class="content">
                  <slot name="default" :item="sub">
                    <span>{{ sub[attr] }}</span>
                  </slot>
                </div>
              </div>
            </li>
          </ul>
        </div>
      </li>
    </ul>
  </div>
</template>

<script>
import { computed, reactive } from 'vue'

export default {
  name: 'SortableTree',

  props: {
    data: { type: Object, default: null },
    attr: { type: String, default: 'name' },
    closeStateKey: { type: String, default: '' },
    childrenAttr: { type: String, default: 'children' },
    mixinParentKey: { type: String, default: '' },
    draggable: { type: Boolean, default: true },
    parentData: { type: Object, default: null },
    idx: { type: Number, default: -1 },
    dragInfo: { type: Object, default: null },
  },

  emits: ['changePosition'],

  setup(props, { emit }) {
    const dragObj = reactive(props.dragInfo ? { ...props.dragInfo } : { data: null, vm: null, vmIdx: -1, parentData: null })

    const showChildren = computed(() => {
      return hasChildren(props.data) && (!props.closeStateKey || !props.data[props.closeStateKey])
    })

    function hasChildren(item) {
      return item && item[props.childrenAttr] && item[props.childrenAttr].length
    }

    function childChildren(item) {
      const children = item[props.childrenAttr]
      const result = []
      if (children && children.length) {
        children.forEach(child => result.push({ _replaceLi_: true, _key: dragObj.data === child ? 'drag-' : 'before-' + (child.pageID || child.pageID || Math.random()) }, child))
        result.push({ _replaceLi_: true, _key: 'after-' + Math.random() })
      }
      result.forEach((r, i) => { if (!r._key) r._key = i })
      return result
    }

    const computedChildren = computed(() => {
      const children = props.data ? props.data[props.childrenAttr] : null
      if (!children || !children.length) return []
      const result = []
      children.forEach(child => result.push({ _replaceLi_: true, _key: 'before-' + (child.pageID || child.pageID || Math.random()), _idx: result.length }, child))
      result.push({ _replaceLi_: true, _key: 'after-' + Math.random() })
      result.forEach((r, i) => { if (!r._key) r._key = i })
      return result
    })

    function liClass(item) {
      return {
        'parent-li': hasChildren(item),
        'exist-li': !item._replaceLi_,
        'blank-li': !!item._replaceLi_,
      }
    }

    function isParent() {
      return props.data === dragObj.parentData
    }

    function isNextToMe() {
      return props.parentData === dragObj.parentData && Math.abs(props.idx - dragObj.vmIdx) === 1
    }

    function isMeOrMyAncestor() {
      let el = document.querySelector('.sortable-tree.draging')
      if (!el) return false
      let parent = el.closest('ul')
      while (parent) {
        if (parent.contains(document.currentScript?.parentNode)) return true
        parent = parent.parentElement?.closest('ul')
      }
      return false
    }

    function isAllowToDrop() {
      return !(isNextToMe() || isParent() || isMeOrMyAncestor())
    }

    function dragStart(event) {
      if (props.data._replaceLi_) return event.preventDefault()
      event.dataTransfer?.setData('text/plain', '')
      dragObj.data = props.data
      dragObj.vm = event.currentTarget
      dragObj.vmIdx = props.idx
      dragObj.parentData = props.parentData
      dragObj.pastIdx = (props.idx - 1) / 2
      event.currentTarget.classList.add('draging')
    }

    function dragEnter() {
      if (dragObj.vm) dragObj.vm.classList.add('draging')
      if (!isAllowToDrop()) return
      event?.currentTarget?.classList.add('droper')
    }

    function dragLeave() {
      event?.currentTarget?.classList.remove('droper')
    }

    function drop(event) {
      if (dragObj.vm) dragObj.vm.classList.remove('draging')
      event.currentTarget.classList.remove('droper')
      if (!isAllowToDrop()) return

      const index = dragObj.parentData[props.childrenAttr].indexOf(dragObj.data)
      dragObj.parentData[props.childrenAttr].splice(index, 1)

      let afterParent = props.parentData
      if (props.data._replaceLi_) {
        if (dragObj.parentData === props.parentData) {
          const changedIdx = props.idx / 2
          props.parentData[props.childrenAttr].splice(index > changedIdx ? changedIdx : changedIdx - 1, 0, dragObj.data)
        } else {
          props.parentData[props.childrenAttr].splice(props.idx / 2, 0, dragObj.data)
        }
      } else {
        afterParent = props.data
        if (!props.data[props.childrenAttr]) {
          props.data[props.childrenAttr] = []
        }
        props.data[props.childrenAttr].push(dragObj.data)
      }

      emit('changePosition', {
        beforeParent: dragObj.parentData,
        data: dragObj.data,
        afterParent,
        beforeIndex: dragObj.pastIdx,
      })
    }

    function dragEnd() {
      document.querySelectorAll('.draging').forEach(el => el.classList.remove('draging'))
      document.querySelectorAll('.droper').forEach(el => el.classList.remove('droper'))
    }

    if (props.mixinParentKey) {
      if (props.data && !props.data[props.mixinParentKey]) {
        props.data[props.mixinParentKey] = props.parentData
      }
    }

    return { dragObj, showChildren, computedChildren, hasChildren, childChildren, liClass, dragStart, dragEnter, dragLeave, drop, dragEnd }
  },
}
</script>

<style lang="scss" scoped>
$content-height: 30px;
$blank-li-height: 5px;

.sortable-tree {
  font-size: 16px;
  min-height: $blank-li-height;

  .content {
    height: $content-height;
    line-height: $content-height;
    user-select: none;
  }

  .blank-placeholder {
    min-height: 0;
    .content { width: 0; height: 0; overflow: hidden; }
  }

  .blank-li {
    .content {
      width: 0;
      height: 0;
      overflow: hidden;
    }
  }

  ul, li {
    margin: 0;
    padding: 0;
  }

  ul {
    position: relative;
    display: list-item;
    list-style: none;
    &:empty { width: 0; height: 0; }
  }

  li {
    position: relative;
    padding-left: 24px;
  }
}

.sortable-tree {
  li {
    position: relative;

    &:before, &:after {
      position: absolute;
      content: '';
    }
    &:before {
      width: 24px;
      height: 100%;
      left: 0;
      top: $content-height / -2;
      border-left: 1px solid #999;
    }
    &:after {
      width: 24px;
      height: $content-height;
      top: $content-height / 2;
      left: 0;
      border-top: 1px solid #999;
    }

    &.parent-li:nth-last-child(2):before {
      width: 24px;
      height: $content-height;
      left: 0;
      top: $content-height / -2;
      border-left: 1px solid #999;
    }

    &.blank-li {
      margin: 0;
      padding: 0;
      width: 100%;
      height: $blank-li-height;

      &:after { width: 0; }
      &:last-child { height: 0; }
    }
  }
}

.draging { background: #EFEEEF; }
.droper { background: lightgreen; }
</style>
