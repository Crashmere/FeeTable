let locks = 0;
let restore: (() => void) | undefined;

export function lockPageScroll() {
  if (locks++ === 0) {
    const { body, documentElement: root } = document;
    const x = window.scrollX,
      y = window.scrollY;
    const bodyStyle = body.getAttribute("style");
    const rootStyle = root.getAttribute("style");
    const scrollbar = window.innerWidth - root.clientWidth;
    const padding = parseFloat(getComputedStyle(body).paddingRight);
    Object.assign(body.style, {
      position: "fixed",
      top: `-${y}px`,
      left: `-${x}px`,
      width: "100%",
      overflow: "hidden",
      paddingRight: `${padding + scrollbar}px`,
    });
    root.style.overflow = "hidden";
    root.style.overscrollBehavior = "none";
    restore = () => {
      if (bodyStyle === null) body.removeAttribute("style");
      else body.setAttribute("style", bodyStyle);
      if (rootStyle === null) root.removeAttribute("style");
      else root.setAttribute("style", rootStyle);
      window.scrollTo({ left: x, top: y, behavior: "instant" });
    };
  }
  let released = false;
  return () => {
    if (released) return;
    released = true;
    if (--locks === 0) {
      restore?.();
      restore = undefined;
    }
  };
}
