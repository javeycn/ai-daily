"use client";

import { useState, useEffect, useRef, useCallback, MutableRefObject } from "react";

/**
 * useLazyRender 使用 IntersectionObserver 实现懒渲染。
 * 当目标元素即将进入视口时才标记为可见，组件可据此决定是否渲染真实内容。
 *
 * @param rootMargin 预加载距离，默认 "100px"（即距离视口 100px 时就触发）
 * @returns [ref, isVisible] — ref 绑定到占位容器，isVisible 表示是否应渲染真实内容
 */
export function useLazyRender<T extends HTMLElement = HTMLDivElement>(
  rootMargin = "100px",
): [MutableRefObject<T | null>, boolean] {
  const ref = useRef<T | null>(null);
  const observerRef = useRef<IntersectionObserver | null>(null);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (!el || isVisible) return;

    observerRef.current = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
          observerRef.current?.disconnect();
        }
      },
      { rootMargin },
    );

    observerRef.current.observe(el);
    return () => observerRef.current?.disconnect();
  }, [rootMargin, isVisible]);

  return [ref, isVisible];
}

/**
 * useIsMobile 判断当前视口是否为移动端（宽度 < 640px，对应 Tailwind 的 sm 断点）。
 * 服务端渲染时默认返回 false，客户端 mount 后取实际值。
 */
export function useIsMobile(breakpoint = 640): boolean {
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    const check = () => setIsMobile(window.innerWidth < breakpoint);
    check();

    const mql = window.matchMedia(`(max-width: ${breakpoint - 1}px)`);
    const handler = (e: MediaQueryListEvent) => setIsMobile(e.matches);
    mql.addEventListener("change", handler);
    return () => mql.removeEventListener("change", handler);
  }, [breakpoint]);

  return isMobile;
}
