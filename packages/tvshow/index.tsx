import { Router } from "design-system/components/router";
import { Home } from "./components/home";
import { Add } from "./components/add";

const routes = [
  {
    pattern: new URLPattern({ pathname: "/tvshow" }),
    component: Home,
  },
  {
    pattern: new URLPattern({ pathname: "/tvshow/add" }),
    component: Add,
  },
];

export function TvShow() {
  return <Router routes={routes} />;
}
