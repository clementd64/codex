import { AuthService, createUnauthenticatedClient } from "proto";

export interface Credentials {
  token: string;
  expiration: bigint;
}

interface OIDCConfig {
  authorization_endpoint: string;
  issuer: string;
  token_endpoint: string;
}

interface OIDCTokens {
  id_token: string;
  refresh_token?: string;
}

const OIDC_STATE_KEY = "oidc_state";
const OIDC_REFRESH_TOKEN_KEY = "oidc_refresh_token";

export function isAuthenticated(credentials: Credentials): boolean {
  return isTokenValid(credentials) || Boolean(localStorage.getItem(OIDC_REFRESH_TOKEN_KEY));
}

export function isTokenValid(credentials: Credentials): boolean {
  return (
    Boolean(credentials.token) && credentials.expiration > BigInt(Math.floor(Date.now() / 1000))
  );
}

export function clearRefreshToken(): void {
  localStorage.removeItem(OIDC_REFRESH_TOKEN_KEY);
}

async function exchangeTokens(tokens: OIDCTokens): Promise<Credentials> {
  const response = await createUnauthenticatedClient(AuthService).token({ token: tokens.id_token });
  if (tokens.refresh_token) {
    localStorage.setItem(OIDC_REFRESH_TOKEN_KEY, tokens.refresh_token);
  }
  return { token: response.token, expiration: response.expiration };
}

function random() {
  return crypto
    .getRandomValues(new Uint8Array(48))
    .toBase64({ alphabet: "base64url", omitPadding: true });
}

async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest("SHA-256", data);
  return new Uint8Array(digest).toBase64({
    alphabet: "base64url",
    omitPadding: true,
  });
}

async function fetchOIDCConfig(): Promise<OIDCConfig> {
  const response = await fetch(
    `${import.meta.env.VITE_OIDC_DISCOVERY_ENDPOINT}/.well-known/openid-configuration`,
  );
  if (!response.ok) {
    throw new Error(`Failed to fetch OIDC configuration: ${response.statusText}`);
  }
  return response.json();
}

export async function authorize() {
  clearRefreshToken();
  const config = await fetchOIDCConfig();
  const state = random();
  const verifier = random();
  const challenge = await generateCodeChallenge(verifier);

  sessionStorage.setItem(
    OIDC_STATE_KEY,
    JSON.stringify({ issuer: config.issuer, state, verifier }),
  );

  const params = new URLSearchParams({
    client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
    response_type: "code",
    scope: "openid offline_access",
    redirect_uri: window.location.origin + "/login/callback",
    state,
    code_challenge: challenge,
    code_challenge_method: "S256",
  });

  window.location.href = `${config.authorization_endpoint}?${params.toString()}`;
}

export async function login(code: string, state: string): Promise<Credentials> {
  const stored = JSON.parse(sessionStorage.getItem(OIDC_STATE_KEY) ?? "{}");
  sessionStorage.removeItem(OIDC_STATE_KEY);

  if (stored.state !== state) {
    throw new Error("Invalid state");
  }

  if (!stored.verifier) {
    throw new Error("Missing code verifier");
  }

  const config = await fetchOIDCConfig();

  const tokens = await fetch(config.token_endpoint, {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body: new URLSearchParams({
      client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
      grant_type: "authorization_code",
      code,
      redirect_uri: window.location.origin + "/login/callback",
      code_verifier: stored.verifier,
    }),
  }).then((response) => {
    if (!response.ok) {
      throw new Error(`Failed to exchange code for token: ${response.statusText}`);
    }
    return response.json() as Promise<OIDCTokens>;
  });

  return exchangeTokens(tokens);
}

let pendingRefresh: Promise<Credentials> | undefined;
export function refresh(): Promise<Credentials> {
  if (pendingRefresh) return pendingRefresh;

  pendingRefresh = (async () => {
    const refreshToken = localStorage.getItem(OIDC_REFRESH_TOKEN_KEY);
    if (!refreshToken) throw new Error("No refresh token available");

    const config = await fetchOIDCConfig();
    const response = await fetch(config.token_endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
        grant_type: "refresh_token",
        refresh_token: refreshToken,
      }),
    });
    if (!response.ok) {
      throw new Error(`Failed to refresh token: ${response.statusText}`);
    }
    const tokens = (await response.json()) as OIDCTokens;
    return exchangeTokens(tokens);
  })().finally(() => {
    pendingRefresh = undefined;
  });

  return pendingRefresh;
}
