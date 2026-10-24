import { goto, useLocation } from "design-system/components/router";
import { MenuActionProvider, useMenuActions } from "design-system/components/menuAction";
import { type ComponentType } from "preact";

export interface MenuEntry {
  name: string;
  icon?: string;
  path: string;
  pattern?: URLPattern;
}

export function Layout({
  children,
  menu,
}: {
  children: preact.ComponentChildren;
  menu: MenuEntry[];
}) {
  const url = useLocation();
  const menuActions = useMenuActions();

  return (
    <div class="grid grid-cols-[auto_1fr] gap-4 p-4 h-dvh bg-app">
      <ul class="menu flex h-full bg-menu text-base-content rounded-box w-60 px-3 py-6">
        <li class="pl-4 pb-4 text-2xl font-bold">Codex</li>
        {menu.map((m) => (
          <li>
            <button
              type="button"
              onClick={() => goto(m.path)}
              class={m.pattern?.exec(url.href) ? "menu-active" : ""}
            >
              {m.icon && <span class={m.icon}></span>}
              <span>{m.name}</span>
            </button>
          </li>
        ))}
        <div class="h-8"></div>
        {menuActions.map((action) => (
          <li>
            <button
              type="button"
              onClick={action.onClick}
              disabled={action.disabled}
              class={`btn ${action.class ?? ""}`}
            >
              {action.icon && <span class={action.icon}></span>}
              <span>{action.name}</span>
            </button>
          </li>
        ))}
        <div class="grow"></div>
        <li>
          <a>
            <span class="icon-[mdi--user]"></span>
            <span>User</span>
          </a>
        </li>
      </ul>
      <div class="h-full overflow-auto">{children}</div>
    </div>
  );
}

export function LayoutMixin(Component: ComponentType, menu: MenuEntry[]) {
  return () => (
    <MenuActionProvider>
      <Layout menu={menu}>
        <Component />
      </Layout>
    </MenuActionProvider>
  );
}
