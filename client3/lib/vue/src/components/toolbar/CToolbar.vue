<template>
  <div class="c-toolbar d-flex flex-column flex-sm-row bg-white gap-1 flex-nowrap overflow-auto">
    <div
      class="d-flex align-items-center gap-1 flex-nowrap flex-shrink-0"
    >
      <slot name="start" />
    </div>

    <div
      class="d-flex align-items-center gap-1 flex-nowrap flex-shrink-0"
    >
      <slot name="center" />
    </div>

    <div
      class="d-flex align-items-center gap-1 flex-nowrap flex-shrink-0"
    >
      <slot name="end" />
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped lang="scss">
.c-toolbar {
  z-index: 1036;
  position: relative;
  min-height: 42px;
  padding: 0.75rem;
  flex-shrink: 0;
}

.fill-width > * {
  flex: 1;
}

@media (min-width: 576px) {
  .c-toolbar > div:last-child {
    margin-left: auto !important;
  }
}

// Below sm the root already stacks start/center/end into 3 rows
// (flex-column, see the template) — but each row was still flex-nowrap
// with the *whole bar* sharing one overflow-auto scroller. A wide "end"
// row (Save/Delete/Clone/Edit can be 4-5 btn-lg buttons) made the entire
// stack scroll sideways together, so reaching Save also dragged the Back
// button in row 1 out of view even though row 1 alone fit fine. Letting
// each row wrap its own buttons onto as many lines as it needs keeps
// everything visible with no shared scrollbar to fight.
@media (max-width: 575.98px) {
  .c-toolbar {
    overflow: visible !important;

    > div {
      flex-wrap: wrap !important;
      // Each row's own children (align-items-center) still pack to the
      // start by default once they wrap, which reads as "stuck to the left
      // edge" once a row is centered in a full-width column — center them
      // as a group instead so a lone leftover item (e.g. Save alone on its
      // line) sits in the middle rather than hugging one side.
      justify-content: center;
      row-gap: 0.25rem;
    }

    // The start slot holds the back/close button — a navigation anchor, not
    // an action group like center/end — so it stays pinned to the left edge
    // (matching where a back control is expected) instead of centering.
    > div:first-child {
      justify-content: flex-start;
    }

    // btn-lg is sized for a spacious desktop bar; on a phone-width toolbar
    // that's the "too big" bulkiness the wrap wasn't fixing — bring it down
    // to roughly the default .btn size instead of the lg one.
    .btn-lg {
      padding: 0.375rem 0.75rem;
      font-size: 1rem;
      border-radius: 0.375rem;
    }
  }
}
</style>
