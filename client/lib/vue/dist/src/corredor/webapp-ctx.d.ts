import { corredor } from '../../../../lib/js/dist';
/**
 * Bare-minimum webapp helper
 */
export default class WebappCtx extends corredor.Ctx {
    constructor(args: corredor.BaseArgs);
    /**
     * Clones context and uses new arguments
     */
    withArgs(args: corredor.BaseArgs): WebappCtx;
}
