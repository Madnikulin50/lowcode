declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<Record<string, unknown>, Record<string, unknown>, unknown>
  export default component
}

declare module 'vue' {
  import { apiClients, EventBus } from '@cortezaproject/corteza-js'
  import { Auth } from './plugins/auth'
  import { UIHooks } from './plugins/ui-hooks'
  import { Settings } from './plugins/settings'
  import { ReminderService } from './plugins/reminder'

  interface ComponentCustomProperties {
    $SystemAPI: InstanceType<typeof apiClients.System>;
    $ComposeAPI: InstanceType<typeof apiClients.Compose>;
    $FederationAPI: InstanceType<typeof apiClients.Federation>;
    $AutomationAPI: InstanceType<typeof apiClients.Automation>;
    $UIHooks: UIHooks;
    $EventBus: EventBus;
    $auth: Auth;
    $Settings: Settings;
    $s: (key: string, defaultValue?: unknown) => unknown;
    $Reminder: ReminderService;
    $t: (key: string, options?: Record<string, unknown>) => string;
  }
}

export {}
