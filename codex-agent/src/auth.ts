import { URL } from "url";

interface TokenResponse {
  access_token: string;
  refresh_token?: string;
  expires_in: number;
  token_type: string;
}

interface AuthConfig {
  baseUrl: string;
  realm: string;
  clientId: string;
  username: string;
  password: string;
}

export class KeycloakAuth {
  private config: AuthConfig;
  private accessToken: string | null = null;
  private refreshToken: string | null = null;
  private expiresAt: number = 0;

  constructor(config: AuthConfig) {
    this.config = config;
  }

  async getAccessToken(): Promise<string> {
    // Return cached token if still valid (with 30s buffer)
    if (this.accessToken && this.expiresAt > Date.now() + 30000) {
      return this.accessToken;
    }

    // Try to refresh if we have a refresh token
    if (this.refreshToken) {
      try {
        return await this.refreshAccessToken();
      } catch (error) {
        // Fall through to password grant
      }
    }

    // Get new token using password grant
    return await this.getNewToken();
  }

  private async getNewToken(): Promise<string> {
    const tokenUrl = new URL(
      `/realms/${this.config.realm}/protocol/openid-connect/token`,
      this.config.baseUrl,
    ).toString();

    const params = new URLSearchParams();
    params.append("grant_type", "password");
    params.append("client_id", this.config.clientId);
    params.append("username", this.config.username);
    params.append("password", this.config.password);

    const response = await fetch(tokenUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: params.toString(),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Failed to authenticate with Keycloak: ${error}`);
    }

    const data = (await response.json()) as TokenResponse;
    this.accessToken = data.access_token;
    this.refreshToken = data.refresh_token || null;
    this.expiresAt = Date.now() + data.expires_in * 1000;

    return this.accessToken;
  }

  private async refreshAccessToken(): Promise<string> {
    if (!this.refreshToken) {
      throw new Error("No refresh token available");
    }

    const tokenUrl = new URL(
      `/realms/${this.config.realm}/protocol/openid-connect/token`,
      this.config.baseUrl,
    ).toString();

    const params = new URLSearchParams();
    params.append("grant_type", "refresh_token");
    params.append("client_id", this.config.clientId);
    params.append("refresh_token", this.refreshToken);

    const response = await fetch(tokenUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: params.toString(),
    });

    if (!response.ok) {
      this.accessToken = null;
      this.refreshToken = null;
      throw new Error("Token refresh failed");
    }

    const data = (await response.json()) as TokenResponse;
    this.accessToken = data.access_token;
    this.refreshToken = data.refresh_token || this.refreshToken;
    this.expiresAt = Date.now() + data.expires_in * 1000;

    return this.accessToken;
  }
}

export function createAuth(): KeycloakAuth {
  const baseUrl = process.env.KEYCLOAK_BASE_URL || "http://keycloak:8080";
  const realm = process.env.KEYCLOAK_REALM || "ticketing";
  const clientId = process.env.KEYCLOAK_CLIENT_ID || "myclient";
  const username = process.env.KEYCLOAK_USERNAME;
  const password = process.env.KEYCLOAK_PASSWORD;

  if (!username || !password) {
    throw new Error(
      "Missing KEYCLOAK_USERNAME and/or KEYCLOAK_PASSWORD environment variables",
    );
  }

  return new KeycloakAuth({
    baseUrl,
    realm,
    clientId,
    username,
    password,
  });
}
