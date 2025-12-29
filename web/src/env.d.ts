/// <reference path="../.astro/types.d.ts" />

interface ImportMetaEnv {
  readonly PUBLIC_API_URL: string;
  readonly PUBLIC_R2_BASE_URL: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
  readonly DEV: boolean;
}