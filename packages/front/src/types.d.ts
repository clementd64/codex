interface ImportMetaEnv {
  readonly VITE_OIDC_CLIENT_ID: string;
  readonly VITE_OIDC_DISCOVERY_ENDPOINT: string;
  readonly VITE_OIDC_PROVIDER_NAME: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
