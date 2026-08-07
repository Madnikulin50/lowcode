<template>
  <div>
    <div class="mb-3 p-2 mb-0">
      <h5 class="d-flex align-items-center p-2 mb-0">
        {{ $t('help.basic-controls-label') }}
        <font-awesome-icon :icon="['fas', 'mouse']" class="ms-2 ml-1" />
      </h5>
      <table class="table table-fixed">
        <thead class="d-none">
          <tr><th></th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="item in basicControls" :key="item.action">
            <td>{{ item.action }}</td>
            <td>{{ item.shortcut }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div class="mb-3 p-2 mb-0">
      <h5 class="d-flex align-items-center pt-2 px-2 mb-0">
        {{ $t('help.keyboard-shortcuts-label') }}
        <font-awesome-icon :icon="['fas', 'keyboard']" class="ms-2 ml-1" />
      </h5>
      <small class="ms-2 ml-1">{{ $t('help.mac-users') }}</small>
      <table class="table table-fixed">
        <thead class="d-none">
          <tr><th></th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="item in keyboardShortcuts" :key="item.action">
            <td>{{ item.action }}</td>
            <td>{{ item.shortcut }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const basicControls = [
  { action: t('help.basic-controls.action.select'), shortcut: t('help.basic-controls.shortcut.left-click.label') },
  { action: t('help.basic-controls.action.select-multiple'), shortcut: t('help.basic-controls.shortcut.left-click.drag') },
  { action: t('help.basic-controls.action.deselect'), shortcut: t('help.basic-controls.shortcut.left-click.background') },
  { action: t('help.basic-controls.action.pan'), shortcut: t('help.basic-controls.shortcut.pan') },
  { action: t('help.basic-controls.action.zoom-in/out'), shortcut: t('help.basic-controls.shortcut.zoom') },
]

const keyboardShortcuts = [
  { action: t('help.keyboard-shortcuts.action.select-all'), shortcut: 'Ctrl + A' },
  { action: t('help.keyboard-shortcuts.action.add-selection'), shortcut: `Ctrl + ${t('help.keyboard-shortcuts.shortcut.left-mouse-button')}` },
  { action: t('help.keyboard-shortcuts.action.delete-selected-elements'), shortcut: 'Delete or Backspace' },
  { action: t('help.keyboard-shortcuts.action.reset-view-to-start-point'), shortcut: 'Ctrl + Space' },
  { action: t('help.keyboard-shortcuts.action.undo'), shortcut: 'Ctrl + Z' },
  { action: t('help.keyboard-shortcuts.action.redo'), shortcut: 'Ctrl + Shift + Z' },
  { action: t('help.keyboard-shortcuts.action.cut'), shortcut: 'Ctrl + X' },
  { action: t('help.keyboard-shortcuts.action.copy'), shortcut: 'Ctrl + C' },
  { action: t('help.keyboard-shortcuts.action.paste'), shortcut: 'Ctrl + V' },
  { action: t('help.keyboard-shortcuts.action.save'), shortcut: 'Ctrl + S' },
  { action: t('help.keyboard-shortcuts.action.nudge'), shortcut: `${t('help.keyboard-shortcuts.shortcut.arrow-keys')} Shift ${t('help.keyboard-shortcuts.shortcut.adjust-distance')}` },
  { action: t('help.keyboard-shortcuts.action.hide-guides'), shortcut: `${t('help.keyboard-shortcuts.shortcut.hold')} Alt` },
  { action: t('help.keyboard-shortcuts.action.show-help'), shortcut: 'Shift + ?' },
]
</script>
