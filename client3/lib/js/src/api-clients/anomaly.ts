/* eslint-disable padded-blocks */

// Hand-written to mirror the shape of the other api-clients/*.ts files
// (see federation.ts) - the anomaly REST surface (server/anomaly/rest.yaml)
// is small enough that generating it via corteza-api-client.js wasn't worth
// the risk of that codegen pass touching unrelated namespaces too.

import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'

interface KV {
  [header: string]: unknown;
}

interface Headers {
  [header: string]: string;
}

interface Ctor {
  baseURL?: string;
  accessTokenFn?: () => string | undefined;
  headers?: Headers;
}

interface CortezaResponse {
  error?: string;
  response?: unknown;
}

function stdResolve (response: AxiosResponse<CortezaResponse>): KV|Promise<never> {
  if (response.data.error) {
    return Promise.reject(response.data.error)
  } else {
    return response.data.response as KV
  }
}

export default class Anomaly {
  protected baseURL?: string;
  protected accessTokenFn?: () => (string | undefined);
  protected headers: Headers = {};

  constructor ({ baseURL, headers, accessTokenFn }: Ctor) {
    this.baseURL = baseURL
    this.accessTokenFn = accessTokenFn
    this.headers = {
      'Content-Type': 'application/json',
    }

    this.setHeaders(headers)
  }

  setAccessTokenFn (fn: () => string | undefined): Anomaly {
    this.accessTokenFn = fn
    return this
  }

  setHeaders (headers?: Headers): Anomaly {
    if (typeof headers === 'object') {
      this.headers = headers
    }

    return this
  }

  setHeader (name: string, value: string | undefined): Anomaly {
    if (value === undefined) {
      delete this.headers[name]
    } else {
      this.headers[name] = value
    }

    return this
  }

  api (): AxiosInstance {
    const headers = { ...this.headers }
    const accessToken = this.accessTokenFn ? this.accessTokenFn() : undefined
    if (accessToken) {
      headers.Authorization = 'Bearer ' + accessToken
    }

    return axios.create({
      withCredentials: true,
      baseURL: this.baseURL,
      headers,
    })
  }

  // Search anomaly rules
  async ruleSearch (a: KV, extra: AxiosRequestConfig = {}): Promise<KV> {
    const {
      namespaceID,
      moduleID,
      query,
      enabled,
      limit,
      pageCursor,
      sort,
    } = (a as KV) || {}
    if (!namespaceID) {
      throw Error('field namespaceID is empty')
    }
    const cfg: AxiosRequestConfig = {
      ...extra,
      method: 'get',
      url: this.ruleSearchEndpoint({ namespaceID }),
    }
    cfg.params = {
      moduleID,
      query,
      enabled,
      limit,
      pageCursor,
      sort,
    }

    return this.api().request(cfg).then(result => stdResolve(result))
  }

  ruleSearchEndpoint (a: KV): string {
    const { namespaceID } = a || {}
    return `/namespace/${namespaceID}/anomaly/rule/`
  }

  // Start watching a module field for anomalies
  async ruleCreate (a: KV, extra: AxiosRequestConfig = {}): Promise<KV> {
    const {
      namespaceID,
      moduleID,
      field,
      detector,
      threshold,
      enabled,
      params,
    } = (a as KV) || {}
    if (!namespaceID) {
      throw Error('field namespaceID is empty')
    }
    if (!moduleID) {
      throw Error('field moduleID is empty')
    }
    if (!field) {
      throw Error('field field is empty')
    }
    if (!detector) {
      throw Error('field detector is empty')
    }
    if (threshold === undefined || threshold === null) {
      throw Error('field threshold is empty')
    }
    const cfg: AxiosRequestConfig = {
      ...extra,
      method: 'post',
      url: this.ruleCreateEndpoint({ namespaceID }),
    }
    cfg.data = {
      moduleID,
      field,
      detector,
      threshold,
      enabled,
      params,
    }
    return this.api().request(cfg).then(result => stdResolve(result))
  }

  ruleCreateEndpoint (a: KV): string {
    const { namespaceID } = a || {}
    return `/namespace/${namespaceID}/anomaly/rule/`
  }

  // Update an anomaly rule
  async ruleUpdate (a: KV, extra: AxiosRequestConfig = {}): Promise<KV> {
    const {
      namespaceID,
      ruleID,
      detector,
      threshold,
      enabled,
      params,
    } = (a as KV) || {}
    if (!namespaceID) {
      throw Error('field namespaceID is empty')
    }
    if (!ruleID) {
      throw Error('field ruleID is empty')
    }
    const cfg: AxiosRequestConfig = {
      ...extra,
      method: 'post',
      url: this.ruleUpdateEndpoint({ namespaceID, ruleID }),
    }
    cfg.data = {
      detector,
      threshold,
      enabled,
      params,
    }
    return this.api().request(cfg).then(result => stdResolve(result))
  }

  ruleUpdateEndpoint (a: KV): string {
    const { namespaceID, ruleID } = a || {}
    return `/namespace/${namespaceID}/anomaly/rule/${ruleID}`
  }

  // Stop watching a module field
  async ruleDelete (a: KV, extra: AxiosRequestConfig = {}): Promise<KV> {
    const { namespaceID, ruleID } = (a as KV) || {}
    if (!namespaceID) {
      throw Error('field namespaceID is empty')
    }
    if (!ruleID) {
      throw Error('field ruleID is empty')
    }
    const cfg: AxiosRequestConfig = {
      ...extra,
      method: 'delete',
      url: this.ruleDeleteEndpoint({ namespaceID, ruleID }),
    }

    return this.api().request(cfg).then(result => stdResolve(result))
  }

  ruleDeleteEndpoint (a: KV): string {
    const { namespaceID, ruleID } = a || {}
    return `/namespace/${namespaceID}/anomaly/rule/${ruleID}`
  }

  // Search anomaly findings
  async findingSearch (a: KV, extra: AxiosRequestConfig = {}): Promise<KV> {
    const {
      namespaceID,
      moduleID,
      recordID,
      status,
      severity,
      limit,
      pageCursor,
      sort,
    } = (a as KV) || {}
    if (!namespaceID) {
      throw Error('field namespaceID is empty')
    }
    const cfg: AxiosRequestConfig = {
      ...extra,
      method: 'get',
      url: this.findingSearchEndpoint({ namespaceID }),
    }
    cfg.params = {
      moduleID,
      recordID,
      status,
      severity,
      limit,
      pageCursor,
      sort,
    }

    return this.api().request(cfg).then(result => stdResolve(result))
  }

  findingSearchEndpoint (a: KV): string {
    const { namespaceID } = a || {}
    return `/namespace/${namespaceID}/anomaly/`
  }

  // Acknowledge, resolve or dismiss a finding
  async findingUpdateStatus (a: KV, extra: AxiosRequestConfig = {}): Promise<KV> {
    const { namespaceID, findingID, status } = (a as KV) || {}
    if (!namespaceID) {
      throw Error('field namespaceID is empty')
    }
    if (!findingID) {
      throw Error('field findingID is empty')
    }
    if (!status) {
      throw Error('field status is empty')
    }
    const cfg: AxiosRequestConfig = {
      ...extra,
      method: 'patch',
      url: this.findingUpdateStatusEndpoint({ namespaceID, findingID }),
    }
    cfg.data = { status }
    return this.api().request(cfg).then(result => stdResolve(result))
  }

  findingUpdateStatusEndpoint (a: KV): string {
    const { namespaceID, findingID } = a || {}
    return `/namespace/${namespaceID}/anomaly/${findingID}/status`
  }
}
