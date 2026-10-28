export interface ProfileDTO {
  id: string;
  displayName: string;
  isBaseline: boolean;
  hasControls: boolean;
  hasCalib: boolean;
}

export interface ControlDTO {
  name: string;
  displayName: string;
  kind: string;
  key: string;
  button: number;
  buttons: number[];
  deviceName: string;
  axisName: string;
}

export interface ConflictDTO {
  control: string;
  reason: string;
}

export interface BindingInput {
  key?: string;
  button?: number;
  deviceInstanceGUID?: string;
  productGUID?: string;
}

export interface BindingReq {
  control: string;
  binding: BindingInput;
  profileIds: string[];
  force: boolean;
  profileOverrides: { profileId: string; force: boolean }[];
}

export interface PlanItem {
  profileId: string;
  conflicts: ConflictDTO[];
  warning: string;
  diffSummary: string;
  force: boolean;
}

export interface PlanResp {
  plans: PlanItem[];
}

export interface ApplyItem {
  profileId: string;
  backupPath: string;
  diffSummary: string;
}

export interface ApplyResp {
  applied: ApplyItem[];
  skipped: { profileId: string; reason: string }[];
}

export interface DeviceInfo {
  name: string;
  instanceGUID: string;
  productGUID: string;
  buttonCount: number;
}

export interface CaptureResp {
  button: number;
  deviceInstanceGUID: string;
  productGUID: string;
}

export interface CalibReq {
  sourceProfileId: string;
  profileIds: string[];
}

export interface CopyReq {
  sourceProfileId: string;
  controls: string[];
  profileIds: string[];
  force: boolean;
  profileOverrides: { profileId: string; force: boolean }[];
}

// initToken resolves the session token from a query string: an explicit
// ?token= wins; otherwise dev mode falls back to "dev"; production returns
// null. Pure, for testability.
export function initToken(search: string, isDev: boolean): string | null {
  const t = new URLSearchParams(search).get('token');
  if (t) return t;
  if (isDev) return 'dev';
  return null;
}

export const token = {
  init(): void {
    const t = initToken(window.location.search, import.meta.env.DEV);
    if (t !== null) {
      sessionStorage.setItem('token', t);
    } else {
      sessionStorage.removeItem('token');
    }
  },
  get(): string | null {
    return sessionStorage.getItem('token');
  },
};

// buildTokenUrl appends ?token= to an API path.
export function buildTokenUrl(path: string): string {
  const t = token.get();
  if (!t) return path;
  const sep = path.includes('?') ? '&' : '?';
  return `${path}${sep}token=${encodeURIComponent(t)}`;
}

// keyToByte maps a keydown event.key to the ASCII byte iRacing stores.
// Only single printable A-Z/0-9 are capturable; everything else is null.
export function keyToByte(key: string): number | null {
  if (key.length !== 1) return null;
  const c = key.charCodeAt(0);
  if (c >= 0x30 && c <= 0x39) return c; // 0-9
  if (c >= 0x41 && c <= 0x5a) return c; // A-Z
  if (c >= 0x61 && c <= 0x7a) return c - 0x20; // a-z -> A-Z
  return null;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(buildTokenUrl(path), {
    method,
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    throw new Error(`${method} ${path} failed: ${res.status}`);
  }
  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}

export const api = {
  fetchProfiles: () => request<ProfileDTO[]>('GET', '/api/profiles'),
  fetchControls: (profileId: string) =>
    request<ControlDTO[]>('GET', `/api/profiles/${encodeURIComponent(profileId)}/controls`),
  planBinding: (req: BindingReq) => request<PlanResp>('POST', '/api/plan/binding', req),
  applyBinding: (req: BindingReq) => request<ApplyResp>('POST', '/api/apply/binding', req),
  planCopy: (req: CopyReq) => request<PlanResp>('POST', '/api/plan/copy', req),
  applyCopy: (req: CopyReq) => request<ApplyResp>('POST', '/api/apply/copy', req),
  planCalib: (req: CalibReq) => request<PlanResp>('POST', '/api/plan/calib', req),
  applyCalib: (req: CalibReq) => request<ApplyResp>('POST', '/api/apply/calib', req),
  fetchInputDevices: () => request<DeviceInfo[]>('GET', '/api/input/devices'),
  capture: (timeoutMs: number) =>
    request<CaptureResp>('POST', '/api/input/capture', { timeoutMs }),
  saveSettings: (profilesDir: string) =>
    request<{ profilesDir: string }>('POST', '/api/settings', { profilesDir }),
};
