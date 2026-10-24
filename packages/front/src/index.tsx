import { render } from "preact";
import { Router } from "design-system/components/router";
import { AuthProvider } from "design-system/components/authprovider";
import { Home } from "./components/home";
import { LayoutMixin } from "./components/layout";
import { TvShow } from "../../tvshow";

const menu = [
  {
    name: "TV Show",
    icon: "icon-[mdi--remote-tv]",
    path: "/tvshow",
    pattern: new URLPattern({ pathname: "/tvshow/*?" }),
  },
];

const routes = [
  {
    pattern: new URLPattern({ pathname: "/tvshow/*?" }),
    component: LayoutMixin(TvShow, menu),
  },
  {
    pattern: new URLPattern({ pathname: "/" }),
    component: LayoutMixin(Home, menu),
  },
];

function App() {
  return (
    <AuthProvider>
      <Router routes={routes} />
    </AuthProvider>
  );
}

render(<App />, document.querySelector("body") as Element);
