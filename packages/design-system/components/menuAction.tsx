import { createContext, type ComponentChildren } from "preact";
import { useCallback, useContext, useLayoutEffect, useState } from "preact/hooks";

export interface MenuAction {
  name: string;
  icon?: string;
  class?: string;
  onClick: () => void;
  disabled?: boolean;
}

type RegisterMenuAction = (actions: readonly MenuAction[]) => () => void;

const MenuActionRegistration = createContext<RegisterMenuAction | null>(null);
const MenuActions = createContext<readonly MenuAction[]>([]);

export function MenuActionProvider({ children }: { children: ComponentChildren }) {
  const [registrations, setRegistrations] = useState(new Map<symbol, readonly MenuAction[]>());
  const register = useCallback<RegisterMenuAction>((actions) => {
    const id = Symbol();
    setRegistrations((current) => new Map(current).set(id, actions));

    return () => {
      setRegistrations((current) => {
        const next = new Map(current);
        next.delete(id);
        return next;
      });
    };
  }, []);

  return (
    <MenuActionRegistration.Provider value={register}>
      <MenuActions.Provider value={Array.from(registrations.values()).flat()}>
        {children}
      </MenuActions.Provider>
    </MenuActionRegistration.Provider>
  );
}

export function useMenuAction(actions: readonly MenuAction[], deps: readonly unknown[]) {
  const register = useContext(MenuActionRegistration);
  if (!register) throw new Error("useMenuAction must be used within MenuActionProvider");

  useLayoutEffect(() => register(actions), [register, ...deps]);
}

export function useMenuActions() {
  return useContext(MenuActions);
}
