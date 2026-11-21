<template>
  <main>
    <div class="logo">
      <img src="/applications/video-conference.png">
    </div>
    <div id="roomselection">
      <span>{{ t('jitsi.toStart') }}</span>
      <input
        id="roomInputField"
        v-model="roomName"
        type="text"
        :placeholder="t('jitsi.roomName')"
      >

      <button
        data-test-id="button-create-room"
        :disabled="jitsi || (cleanup(roomName).length === 0)"
        @click="onCreate"
      >
        {{ t('jitsi.create') }}
      </button>

      <div v-show="jitsi" ref="jitsiInterface" class="jitsiInterface" />
    </div>
  </main>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'app', keyPrefix: 'jitsi' } })
import { ref, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuth } from 'corteza-lib/vue/dist'

const script = document.createElement('script')
script.src = 'https://meet.jit.si/external_api.js'
document.head.appendChild(script)

const domain = 'meet.jit.si'

const { t } = useI18n()
const { auth } = useAuth()

const roomName = ref('')
const jitsi = ref(null)
const jitsiInterface = ref(null)

function dispose() {
  if (jitsi.value) {
    jitsi.value.dispose()
    jitsi.value = null
  }
}

function cleanup(str) {
  return str.replace(/[^a-z0-9+]+/gi, '')
}

function onJoin() {
  open({
    roomName: roomName.value,
    userDisplayName: auth.user?.name || auth.user?.email,
  })
}

function onCreate() {
  open({
    roomName: roomName.value,
    userDisplayName: auth.user?.name || auth.user?.email,
  })
}

function onClose() {
  dispose()
}

function removeJitsiAfterHangup() {
  dispose()
}

function open({ roomName: rn, userDisplayName } = {}) {
  dispose()

  /* eslint-disable-next-line no-undef */
  const JitsiMeetExternalAPI = window.JitsiMeetExternalAPI
  if (!JitsiMeetExternalAPI) return

  jitsi.value = new JitsiMeetExternalAPI(domain, {
    roomName: `crust_${cleanup(rn || 'unnamed')}`,
    width: '100%',
    height: '100%',
    parentNode: jitsiInterface.value,
    interfaceConfigOverwrite: {
      DEFAULT_BACKGROUND: '#232323',
      SHOW_JITSI_WATERMARK: true,
      SHOW_WATERMARK_FOR_GUESTS: false,
      SHOW_BRAND_WATERMARK: false,
      BRAND_WATERMARK_LINK: '',
      SHOW_POWERED_BY: false,
      DEFAULT_REMOTE_DISPLAY_NAME: t('jitsi.defaultRemoteDisplayName'),
      DEFAULT_LOCAL_DISPLAY_NAME: userDisplayName || t('jitsi.defaultLocalDisplayName'),
      TOOLBAR_BUTTONS: [
        'microphone', 'camera', 'closedcaptions', 'desktop', 'fullscreen',
        'fodeviceselection', 'hangup', 'profile', 'info', 'recording',
        'settings', 'tileview', 'videoquality', 'filmstrip', 'invite', 'shortcuts',
      ],
      SETTINGS_SECTIONS: ['devices', 'language', 'moderator', 'profile', 'calendar'],
    },
  })

  jitsi.value.addEventListeners({
    readyToClose: removeJitsiAfterHangup,
  })

  window.jitsi = jitsi.value
}

onBeforeUnmount(() => {
  dispose()
})
</script>

<style lang="scss" scoped>
main {
  overflow: auto;
  height: 100vh;

  .logo {
    text-align: center;
    margin-top: 4rem;
    img {
      max-width: 200px;
    }
  }

  .jitsiInterface {
    position: absolute;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    display: flex;
    flex-direction: column;
    background: #232323;

    & > iframe {
      flex: 1 1 auto;
    }
  }

  #roomselection {
    max-width: 400px;
    margin: 50px auto;
    padding: 50px;
    background: var(--white);
  }

  input {
    height: 30px;
    width: 100%;
    border: 1px solid var(--secondary);
    padding-left: 10px;
    font-size: 14px;
    display: block;
    margin-top: 10px;
    box-sizing: border-box;
  }

  select {
    height: 30px;
    width: 100%;
    margin-top: 10px;
    background: transparent;
    padding-left: 10px;
    font-size: 14px;
    border-radius: 0;
    appearance: none;
    border: 1px solid var(--secondary);
  }

  #roomdropdown::after {
    border: 4px dashed transparent;
    border-top: 4px solid var(--secondary);
    content: "";
    display: inline-block;
    float: right;
    margin-right: 10px;
    margin-top: -15px;
  }

  select:focus,
  input:focus {
    outline: none;
  }

  button {
    cursor: pointer;
    background: transparent;
    color: var(--primary);
    font-size: 14px;
    line-height: 38px;
    text-decoration: none;
    display: block;
    width: 150px;
    text-align: center;
    height: 40px;
    margin: 20px auto 0;
    transition: color .2s,background-color .2s;
    border: 1px solid var(--primary);
    &:hover {
      border: 1px solid var(--primary);
      background: var(--primary);
      color: var(--white);
    }
    &:disabled {
      cursor: not-allowed;
      color: var(--secondary);
      border-color: var(--secondary);
      &:hover {
        background: transparent;
      }
    }
  }

  h4 {
    display: flex;
    width: 100%;
    justify-content: center;
    align-items: center;
    text-align: center;
    margin: 30px 0;
    color: var(--secondary);
    &:before,
    &:after {
      content: '';
      border-top: 1px solid var(--secondary);
      margin: 0 20px 0 0;
      flex: 1 0 20px;
    }
    &:after {
      margin: 0 0 0 20px;
    }
  }
}
</style>
