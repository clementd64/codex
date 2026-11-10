import { Router } from "design-system/components/router";
import { Home } from "./components/home";
import { Search } from "./components/search";

const routes = [
  {
    pattern: new URLPattern({ pathname: "/game" }),
    component: Home,
  },
  {
    pattern: new URLPattern({ pathname: "/game/search" }),
    component: Search,
  },
];

export function Game() {
  return <Router routes={routes} />;
}
