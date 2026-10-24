export function countryCodeToFlag(code: string) {
  return /^[a-z]{2}$/i.test(code)
    ? String.fromCodePoint(...[...code.toUpperCase()].map((c) => 127397 + c.charCodeAt(0)))
    : "";
}
