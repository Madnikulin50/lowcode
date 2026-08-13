import { ref, onMounted, onBeforeUnmount } from 'vue'

export function useRecordLayout (containerRef) {
  const columnWrapClass = ref('')
  let resizeObserver = null

  const breakpoints = { xs: 576, md: 768, lg: 992, xl: 1200 }
  const columnClasses = { xs: 'col-12', md: 'col-6', lg: 'col-4', xl: 'col-3' }

  function applyColumnClasses (width) {
    let columnClass
    if (width <= breakpoints.xs) {
      columnClass = columnClasses.xs
    } else if (width <= breakpoints.md) {
      columnClass = columnClasses.md
    } else if (width <= breakpoints.lg) {
      columnClass = columnClasses.lg
    } else {
      columnClass = columnClasses.xl
    }
    columnWrapClass.value = `field-col ${columnClass}`
  }

  function initObserver (el) {
    if (!el || resizeObserver) return
    resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        applyColumnClasses(entry.contentRect.width)
      }
    })
    resizeObserver.observe(el)
  }

  function destroyObserver () {
    if (resizeObserver) {
      resizeObserver.disconnect()
      resizeObserver = null
    }
  }

  onBeforeUnmount(() => {
    destroyObserver()
  })

  return { columnWrapClass, initObserver, destroyObserver }
}
