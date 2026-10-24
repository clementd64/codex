import { type DescService } from "@bufbuild/protobuf";
import { type Client } from "@connectrpc/connect";
import { useContext, useMemo, useRef } from "preact/hooks";
import { createClient } from "proto";

import { Auth } from "design-system/components/authprovider";
import { clearRefreshToken, isTokenValid, refresh } from "design-system/lib/auth";
import { goto } from "design-system/components/router";

export function useClient<T extends DescService>(service: T): Client<T> {
  const [credentials, setCredentials] = useContext(Auth);
  const current = useRef(credentials);
  current.current = credentials;

  return useMemo(() => {
    return createClient(service, {
      async token() {
        if (!isTokenValid(current.current)) await this.refresh();
        return current.current.token;
      },
      async refresh() {
        try {
          current.current = await refresh();
          setCredentials(current.current);
        } catch (error) {
          this.logout();
          throw error;
        }
      },
      logout() {
        clearRefreshToken();
        current.current = { token: "", expiration: 0n };
        setCredentials(current.current);
        goto("/login", true);
      },
    });
  }, [service, setCredentials]);
}
