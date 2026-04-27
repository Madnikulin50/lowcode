import { PluginFunction } from 'vue';
import { eventbus } from '../../../../lib/js/dist';
export default function (): PluginFunction<Partial<eventbus.Options>>;
