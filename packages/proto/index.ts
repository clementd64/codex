import { type DescService } from "@bufbuild/protobuf";
import { createConnectTransport } from "@connectrpc/connect-web";
import {
  Code,
  ConnectError,
  createClient as createConnectClient,
  type Client,
} from "@connectrpc/connect";

export interface TokenManager {
  token(): Promise<string>;
  refresh(): Promise<void>;
  logout(): void;
}

export function createClient<T extends DescService>(service: T, manager: TokenManager): Client<T> {
  const transport = createConnectTransport({
    baseUrl: "/api",
    interceptors: [
      (next) => async (request) => {
        request.header.set("Authorization", `Bearer ${await manager.token()}`);
        try {
          return await next(request);
        } catch (error) {
          if (
            request.stream ||
            request.signal.aborted ||
            ConnectError.from(error).code !== Code.Unauthenticated
          ) {
            throw error;
          }

          try {
            await manager.refresh();
            request.header.set("Authorization", `Bearer ${await manager.token()}`);
            return await next(request);
          } catch (retryError) {
            // a non auth error should never hit this step
            manager.logout();
            throw retryError;
          }
        }
      },
    ],
  });
  return createConnectClient(service, transport);
}

export function createUnauthenticatedClient<T extends DescService>(service: T): Client<T> {
  return createConnectClient(service, createConnectTransport({ baseUrl: "/api" }));
}

export { AuthService } from "./proto/auth/auth_pb";
export { TvShowService } from "./proto/tvshow/tvshow_pb";
