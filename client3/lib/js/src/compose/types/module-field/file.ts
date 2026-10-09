import { ModuleField, Registry, Options, defaultOptions } from './base'
import { Apply, ApplyWhitelisted } from '../../../cast'

const kind = 'File'

export const modes = [
  // list of attachments, no preview
  'list',
  // list of all images/files, show preview
  'gallery',
]

// Empty string ('') means "inherit the system-wide default" (Compose ›
// Attachments › Default storage, itself "db" unless an admin changes it).
export const storageDrivers = ['', 'db', 'plain', 'minio']

interface FileOptions extends Options {
  allowImages: boolean;
  allowDocuments: boolean;
  maxSize: number;
  mode: string;
  inline: boolean;
  hideFileName: boolean;
  mimetypes?: string;
  // Which objstore.Store backend this field's uploads are saved to.
  // '' = system default, 'db' = stored in the database, 'plain' = local
  // disk, 'minio' = S3-compatible storage.
  storageDriver?: string;
  height?: string;
  width?: string;
  maxHeight?: string;
  maxWidth?: string;
  borderRadius?: string;
  margin?: string;
  backgroundColor?: string;
  clickToView?: boolean;
  enableDownload?: boolean;
  multiDelimiter: string;
  enableWebcam?: boolean
}

const defaults = (): Readonly<FileOptions> => Object.freeze({
  ...defaultOptions(),
  allowImages: true,
  allowDocuments: true,
  maxSize: 0,
  mode: 'list',
  inline: true,
  hideFileName: false,
  mimetypes: '',
  height: '',
  width: '',
  maxHeight: '',
  maxWidth: '',
  borderRadius: '',
  margin: 'auto',
  backgroundColor: '#FFFFFF00',
  clickToView: true,
  enableDownload: true,
  multiDelimiter: '\n',
  enableWebcam: false,
  storageDriver: '',
})

export class ModuleFieldFile extends ModuleField {
  readonly kind = kind

  options: FileOptions = { ...defaults() }

  constructor (i?: Partial<ModuleFieldFile>) {
    super(i)
    this.applyOptions(i?.options)
  }

  applyOptions (o?: Partial<FileOptions>): void {
    if (!o) return
    super.applyOptions(o)

    Apply(this.options, o, Number, 'maxSize')
    Apply(this.options, o, Boolean, 'allowImages', 'allowDocuments', 'inline', 'hideFileName', 'clickToView', 'enableDownload', 'enableWebcam')
    Apply(this.options, o, String, 'mimetypes', 'height', 'width', 'maxHeight', 'maxWidth', 'borderRadius', 'margin', 'backgroundColor')

    // Legacy
    if (o.mode === 'single') {
      o.mode = 'gallery'
    } else if (o.mode === 'grid') {
      o.mode = 'list'
    }

    ApplyWhitelisted(this.options, o, modes, 'mode')
    ApplyWhitelisted(this.options, o, storageDrivers, 'storageDriver')
  }
}

Registry.set(kind, ModuleFieldFile)
