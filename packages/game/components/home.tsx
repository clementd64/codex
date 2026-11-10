import { useMenuAction } from "design-system/components/menuAction";
import { goto } from "design-system/components/router";

export function Home() {
  useMenuAction(
    [
      {
        name: "Search",
        icon: "icon-[mdi--search]",
        class: "btn-primary",
        onClick() {
          goto("/game/search");
        },
      },
    ],
    [],
  );

  return null;
}
