import lodash from 'lodash'
const { merge } = lodash
import { Apply, CortezaID, ISO8601Date, NoID } from '../../cast'
import { PageBlock } from './page-block/base'
import { Button } from './page-block/types'

export type PageLayoutInput = PageLayout | Partial<PageLayout>

interface PageLayoutConfig {
  visibility: Visibility;
  buttons: {
    back: Button;
    delete: Button;
    new: Button;
    clone: Button;
    edit: Button;
    submit: Button;
  };
  actions: Action[];
  // Only used for record pages
  useTitle: boolean;
  validation: Validation;
}

interface RequiredField {
  field: string;
  condition: string;
}

interface Validation {
  requiredFields: RequiredField[];
}

interface Action {
  kind: string;
  enabled: boolean;
  placement: string;
  params: unknown;
  meta: ActionMeta;
}

interface ActionMeta {
  label: string;
  style: {
    variant: string;
  }
}

interface Visibility {
  expression: string;
  roles: string[];
}

interface Meta {
  title: string;
  description: string;
}

export class PageLayout {
  public pageLayoutID = NoID;
  public namespaceID = NoID;
  public pageID = NoID
  public handle = '';

  public weight = 0;

  public blocks: (Partial<PageBlock>)[] = [];

  public config: PageLayoutConfig = {
    visibility: {
      expression: '',
      roles: [],
    },
    buttons: {
      back: { enabled: true },
      delete: { enabled: true },
      clone: { enabled: true },
      new: { enabled: true },
      edit: { enabled: true },
      submit: { enabled: true },
    },
    actions: [],
    useTitle: false,
    validation: {
      requiredFields: [],
    },
  }

  public meta: Meta = {
    title: '',
    description: '',
  };

  public createdAt?: Date = undefined;
  public updatedAt?: Date = undefined;
  public deletedAt?: Date = undefined;

  public ownedBy = NoID;

  constructor (pl?: PageLayoutInput) {
    this.apply(pl)
  }

  apply (pl?: PageLayoutInput): void {
    if (!pl) return

    Apply(this, pl, CortezaID, 'pageLayoutID', 'namespaceID', 'pageID', 'ownedBy')
    Apply(this, pl, String, 'handle')
    Apply(this, pl, Number, 'weight')
    Apply(this, pl, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt')

    // Was (({ blockID, xywh, meta }) => ({ blockID, xywh, meta })) — silently
    // dropped kind/title/description/prompt/options/style on every load into
    // the pageLayout Pinia store, even though the server always sends the
    // full shape (see PageBlock, ./page-block/base.ts). Any page-builder save
    // made from this in-memory state would then round-trip that stripped
    // shape straight back to the server, permanently corrupting the stored
    // layout — the same bug existed server-side in
    // server/compose/types/page_layout.go's PageLayoutBlock (fixed
    // separately) for the exact same reason: a struct/type only wide enough
    // for the 3 fields early code needed, never widened when blocks grew
    // kind/title/options/style.
    this.blocks = (pl.blocks || []).map(b => new PageBlock(b))

    if (pl.meta) {
      this.meta = { ...this.meta, ...pl.meta }
    }

    if (pl.config) {
      this.config = merge({}, this.config, pl.config)
    }
  }

  clone (): PageLayout {
    return new PageLayout(JSON.parse(JSON.stringify(this)))
  }

  addAction () {
    this.config.actions.push({
      kind: 'toLayout',
      placement: 'end',
      enabled: true,
      params: {
        pageLayoutID: '',
      },
      meta: {
        label: '',
        style: {
          variant: 'primary',
        },
      },
    } as Action)
  }

  /**
   * Returns resource ID
   */
  get resourceID (): string {
    return `${this.resourceType}:${this.pageLayoutID}`
  }

  /**
   * Resource type
   */
  get resourceType (): string {
    return 'compose:page-layout'
  }

  export (): PageLayoutInput {
    return {
      blocks: this.blocks,
      meta: this.meta,
    }
  }
}
