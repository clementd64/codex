import { useMenuAction } from "design-system/components/menuAction";
import { goto } from "design-system/components/router";

export function Home() {
  useMenuAction(
    [
      {
        name: "Add",
        icon: "icon-[mdi--add]",
        class: "btn-primary",
        onClick() {
          goto("/game/add");
        },
      },
    ],
    [],
  );

  return null;
}
