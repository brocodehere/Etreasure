export type AuthTokens = {
  accessToken: string;
  refreshToken: string;
};

const ACCESS_KEY = 'etreasure_admin_access';
const REFRESH_KEY = 'etreasure_admin_refresh';

function getStorage(): any {
  try {
    const g: any = globalThis as any;
    return g && g.localStorage ? g.localStorage : null;
  } catch {
    return null;
  }
}

export function saveTokens(tokens: AuthTokens) {
  const store = getStorage();
  if (store) {
    store.setItem(ACCESS_KEY, tokens.accessToken);
    store.setItem(REFRESH_KEY, tokens.refreshToken);
  }
}

export function clearTokens() {
  const store = getStorage();
  if (store) {
    store.removeItem(ACCESS_KEY);
    store.removeItem(REFRESH_KEY);
  }
}

export function getAccessToken(): string | null {
  const store = getStorage();
  return store ? store.getItem(ACCESS_KEY) : null;
}

export function getRefreshToken(): string | null {
  const store = getStorage();
  return store ? store.getItem(REFRESH_KEY) : null;
}
