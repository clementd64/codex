import { Router } from "design-system/components/router";
import { Home } from "./components/home";
import { Add } from "./components/add";

const routes = [
  {
    pattern: new URLPattern({ pathname: "/game" }),
    component: Home,
  },
  {
    pattern: new URLPattern({ pathname: "/game/add" }),
    component: Add,
  },
];

export function Game() {
  return <Router routes={routes} />;
}
