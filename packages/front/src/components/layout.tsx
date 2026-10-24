import { goto, useLocation } from "design-system/components/router";
import { MenuActionProvider, useMenuActions } from "design-system/components/menuAction";
import { type ComponentType } from "preact";

export interface MenuEntry {
  name: string;
  icon: string;
  path: string;
  pattern?: URLPattern;
  bottom?: boolean;
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

  const menuEntries = (bottom = false) =>
    menu
      .filter((m) => Boolean(m?.bottom) === bottom)
      .map((m) => (
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
      ));

  return (
    <div class="grid grid-rows-[1fr_auto] h-dvh bg-app">
      <div class="grid lg:grid-cols-[auto_1fr] overflow-y-auto lg:p-4 lg:gap-4 relative">
        <ul class="menu hidden lg:flex bg-menu text-base-content rounded-box w-60 px-3 py-6">
          <li class="pl-4 pb-4 text-2xl font-bold">Codex</li>
          {menuEntries()}

          {menuActions.length > 0 ? (
            <>
              <div class="h-8"></div>
              {menuActions.map((action) => (
                <li>
                  <button
                    type="button"
                    onClick={action.onClick}
                    class={`btn ${action.class ?? ""}`}
                  >
                    <span class={action.icon}></span>
                    <span>{action.name}</span>
                  </button>
                </li>
              ))}
            </>
          ) : null}

          <div class="grow"></div>
          {menuEntries(true)}
        </ul>

        <div class="overflow-y-auto p-4 lg:p-0">{children}</div>

        {menuActions.length > 0 ? (
          <div class="fab absolute lg:hidden">
            <div tabindex={0} role="button" class="btn btn-lg btn-circle">
              <span class="icon-[mdi--plus]"></span>
            </div>

            <div class="fab-close">
              Close
              <span class="btn btn-circle btn-lg btn-error">
                <span class="icon-[mdi--close]"></span>
              </span>
            </div>

            {menuActions.map((action) => (
              <div>
                {action.name}
                <a onClick={action.onClick} class={`btn btn-lg btn-circle ${action.class ?? ""}`}>
                  <span class={action.icon}></span>
                </a>
              </div>
            ))}
          </div>
        ) : null}
      </div>

      <div class="dock static lg:hidden bg-menu">
        {menu.map((m) => (
          <a onClick={() => goto(m.path)} class={m.pattern?.exec(url.href) ? "dock-active" : ""}>
            {m.icon && <span class={m.icon}></span>}
            <span class="dock-label">{m.name}</span>
          </a>
        ))}
      </div>
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
