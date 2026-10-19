"use client";

import { useCallback, useEffect, useMemo, useRef, useState, type PointerEvent as ReactPointerEvent } from "react";
import { ChevronsDownUp, ChevronsUpDown, Maximize2, Minus, Plus } from "lucide-react";

export interface CategoryMapNode {
  id: string;
  name: string;
  parentId?: string | null;
  productsCount?: number;
  slug?: string;
}

interface CategoryTreeMapProps {
  categories: CategoryMapNode[];
  selectedId?: string | null;
  onSelect: (category: CategoryMapNode) => void;
  emptyLabel: string;
  emptyNodeLabel: string;
  itemsLabel: (count: number) => string;
  hint: string;
  collapseLabel: string;
  expandLabel: string;
  collapseAllLabel: string;
  expandAllLabel: string;
}

interface TreeNode extends CategoryMapNode {
  parentId: string | null;
  children: TreeNode[];
}

interface LayoutNode extends CategoryMapNode {
  depth: number;
  x: number;
  y: number;
  // Products in this category plus all descendants; makes parent nodes show a
  // meaningful total instead of "No products".
  totalProducts: number;
  // Direct children plus all deeper descendants, from the full tree even when
  // this branch is collapsed.
  childCount: number;
  descendantCount: number;
  collapsed: boolean;
}

interface Layout {
  nodes: LayoutNode[];
  width: number;
  height: number;
}

// Nodes are rendered from an animated snapshot rather than directly from the
// layout, so collapsing a branch can visually pull its descendants into the
// parent while every other node keeps its position.
interface AnimatedNode extends LayoutNode {
  displayX: number;
  displayY: number;
  scale: number;
  opacity: number;
  entering: boolean;
  exiting: boolean;
}

interface AnimationStep {
  node: LayoutNode;
  entering: boolean;
  exiting: boolean;
  fromX: number;
  fromY: number;
  fromScale: number;
  fromOpacity: number;
  toX: number;
  toY: number;
  toScale: number;
  toOpacity: number;
}

// Geometry for the tidy tree layout. Nodes are laid out column-per-depth, and
// each parent is vertically centred on its children's span.
const NODE_WIDTH = 190;
const NODE_HEIGHT = 56;
const HORIZONTAL_GAP = 76;
const VERTICAL_GAP = 16;
const ROOT_GAP = 28;
const COLUMN_WIDTH = NODE_WIDTH + HORIZONTAL_GAP;

function buildLayout(categories: CategoryMapNode[]): Layout {
  const byId = new Map<string, TreeNode>();
  for (const category of categories) {
    byId.set(category.id, { ...category, parentId: category.parentId ?? null, children: [] });
  }

  // Break malformed cycles defensively: demote any node whose ancestor chain
  // loops back to itself to a root, so layout can never recurse forever.
  for (const node of byId.values()) {
    const seen = new Set<string>([node.id]);
    let current: TreeNode | undefined = node;
    while (current?.parentId && byId.has(current.parentId)) {
      if (seen.has(current.parentId)) {
        node.parentId = null;
        break;
      }
      seen.add(current.parentId);
      current = byId.get(current.parentId);
    }
  }

  const roots: TreeNode[] = [];
  for (const node of byId.values()) {
    if (node.parentId && byId.has(node.parentId) && node.parentId !== node.id) {
      byId.get(node.parentId)!.children.push(node);
    } else {
      roots.push(node);
    }
  }

  const byName = (a: TreeNode, b: TreeNode) => a.name.localeCompare(b.name);
  roots.sort(byName);
  for (const node of byId.values()) {
    node.children.sort(byName);
  }

  // Rollups are computed over the full tree, regardless of which branches are
  // collapsed, so a collapsed parent still advertises its true totals.
  const totals = new Map<string, number>();
  const descendantCounts = new Map<string, number>();

  const summarize = (node: TreeNode): { total: number; descendants: number } => {
    let total = node.productsCount ?? 0;
    let descendants = 0;
    for (const child of node.children) {
      const summary = summarize(child);
      total += summary.total;
      descendants += summary.descendants + 1;
    }
    totals.set(node.id, total);
    descendantCounts.set(node.id, descendants);
    return { total, descendants };
  };
  roots.forEach(summarize);

  const nodes: LayoutNode[] = [];
  let cursorY = 0;
  let maxDepth = 0;

  const walk = (node: TreeNode, depth: number): number => {
    maxDepth = Math.max(maxDepth, depth);
    const x = depth * COLUMN_WIDTH;

    let centerY: number;
    if (node.children.length === 0) {
      centerY = cursorY + NODE_HEIGHT / 2;
      cursorY += NODE_HEIGHT + VERTICAL_GAP;
    } else {
      const childCenters = node.children.map((child) => walk(child, depth + 1));
      centerY = (childCenters[0] + childCenters[childCenters.length - 1]) / 2;
    }

    const y = centerY - NODE_HEIGHT / 2;
    nodes.push({
      id: node.id,
      name: node.name,
      parentId: node.parentId,
      productsCount: node.productsCount,
      slug: node.slug,
      depth,
      x,
      y,
      totalProducts: totals.get(node.id) ?? 0,
      childCount: node.children.length,
      descendantCount: descendantCounts.get(node.id) ?? 0,
      collapsed: false,
    });

    return centerY;
  };

  roots.forEach((root, index) => {
    walk(root, 0);
    if (index < roots.length - 1) {
      cursorY += ROOT_GAP;
    }
  });

  return {
    nodes,
    width: maxDepth * COLUMN_WIDTH + NODE_WIDTH,
    height: Math.max(cursorY - VERTICAL_GAP, NODE_HEIGHT),
  };
}

/**
 * Positions come from the fully expanded tree, so collapsing or expanding a
 * branch only hides/shows nodes; it never re-packs the rest of the map. Hidden
 * descendants reserve their space until they are shown again.
 */
function visibleLayout(full: Layout, collapsedIds: ReadonlySet<string>): Layout {
  const byId = new Map(full.nodes.map((node) => [node.id, node]));
  const hidden = new Set<string>();

  for (const node of full.nodes) {
    let parentId = node.parentId ?? null;
    const seen = new Set<string>([node.id]);

    while (parentId && !seen.has(parentId)) {
      if (collapsedIds.has(parentId)) {
        hidden.add(node.id);
        break;
      }
      seen.add(parentId);
      parentId = byId.get(parentId)?.parentId ?? null;
    }
  }

  return {
    ...full,
    nodes: full.nodes
      .filter((node) => !hidden.has(node.id))
      .map((node) => ({
        ...node,
        collapsed: collapsedIds.has(node.id) && node.childCount > 0,
      })),
  };
}

const COLLAPSE_DURATION = 380;
const EXIT_SCALE = 0.18;

function toStaticNodes(nodes: LayoutNode[]): AnimatedNode[] {
  return nodes.map((node) => ({
    ...node,
    displayX: node.x,
    displayY: node.y,
    scale: 1,
    opacity: 1,
    entering: false,
    exiting: false,
  }));
}

function easeInOutCubic(t: number): number {
  return t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
}

/**
 * Builds the from/to pairs for a collapse or expand transition.
 *
 * - Nodes present in both layouts glide to their new slots.
 * - Exiting descendants converge on the nearest ancestor that survives the
 *   collapse and shrink to a point, so a subtree visually folds into its parent.
 * - Entering descendants grow out of the nearest ancestor that was visible
 *   before the expansion.
 */
function buildAnimation(previous: AnimatedNode[], target: Layout): AnimationStep[] {
  const previousById = new Map(previous.map((node) => [node.id, node]));
  const targetById = new Map(target.nodes.map((node) => [node.id, node]));
  const ids = new Set<string>([...previousById.keys(), ...targetById.keys()]);
  const steps: AnimationStep[] = [];

  for (const id of ids) {
    const previousNode = previousById.get(id);
    const targetNode = targetById.get(id);

    if (previousNode && targetNode) {
      steps.push({
        node: targetNode,
        entering: false,
        exiting: false,
        fromX: previousNode.displayX,
        fromY: previousNode.displayY,
        fromScale: previousNode.scale,
        fromOpacity: previousNode.opacity,
        toX: targetNode.x,
        toY: targetNode.y,
        toScale: 1,
        toOpacity: 1,
      });
      continue;
    }

    if (targetNode) {
      // Entering: walk up the target tree until we find a node that was already
      // on screen, then emerge from its current position.
      let cursor: LayoutNode | undefined = targetNode;
      let anchor: AnimatedNode | undefined;
      const visited = new Set<string>();
      while (cursor?.parentId && !visited.has(cursor.id)) {
        visited.add(cursor.id);
        const parent = targetById.get(cursor.parentId);
        if (!parent) break;
        anchor = previousById.get(parent.id);
        if (anchor) break;
        cursor = parent;
      }

      steps.push({
        node: targetNode,
        entering: true,
        exiting: false,
        fromX: anchor?.displayX ?? targetNode.x,
        fromY: anchor?.displayY ?? targetNode.y,
        fromScale: EXIT_SCALE,
        fromOpacity: 0,
        toX: targetNode.x,
        toY: targetNode.y,
        toScale: 1,
        toOpacity: 1,
      });
      continue;
    }

    if (previousNode) {
      // Exiting: walk up the previous tree until we find a node that survives
      // the collapse and slide into that node's next position.
      let cursor: AnimatedNode | undefined = previousNode;
      let anchor: LayoutNode | undefined;
      const visited = new Set<string>();
      while (cursor?.parentId && !visited.has(cursor.id)) {
        visited.add(cursor.id);
        anchor = targetById.get(cursor.parentId);
        if (anchor) break;
        cursor = previousById.get(cursor.parentId);
      }

      steps.push({
        node: previousNode,
        entering: false,
        exiting: true,
        fromX: previousNode.displayX,
        fromY: previousNode.displayY,
        fromScale: previousNode.scale,
        fromOpacity: previousNode.opacity,
        toX: anchor?.x ?? previousNode.x,
        toY: anchor?.y ?? previousNode.y,
        toScale: Math.min(EXIT_SCALE, previousNode.scale),
        toOpacity: 0,
      });
    }
  }

  return steps;
}

interface RenderEdge {
  id: string;
  fromX: number;
  fromY: number;
  toX: number;
  toY: number;
  opacity: number;
}

function buildEdges(nodes: AnimatedNode[]): RenderEdge[] {
  const byId = new Map(nodes.map((node) => [node.id, node]));
  const edges: RenderEdge[] = [];

  for (const node of nodes) {
    if (!node.parentId) continue;
    const parent = byId.get(node.parentId);
    if (!parent) continue;

    const opacity = Math.min(parent.opacity, node.opacity);
    if (opacity <= 0.01) continue;

    const parentCenterY = parent.displayY + NODE_HEIGHT / 2;
    const childCenterY = node.displayY + NODE_HEIGHT / 2;
    const fromX = parent.displayX + NODE_WIDTH / 2 + (NODE_WIDTH / 2) * parent.scale;
    const toX = node.displayX + NODE_WIDTH / 2 - (NODE_WIDTH / 2) * node.scale;

    edges.push({
      id: `${parent.id}-${node.id}`,
      fromX,
      fromY: parentCenterY,
      toX,
      toY: childCenterY,
      opacity,
    });
  }

  return edges;
}

interface Viewport {
  x: number;
  y: number;
  scale: number;
}

const MIN_SCALE = 0.25;
const MAX_SCALE = 3;

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

/**
 * Interactive category hierarchy map.
 *
 * Renders an SVG tree with one column per depth level. It is dependency-free:
 * pan uses pointer events, zoom uses a non-passive wheel listener so the page
 * does not scroll while zooming. Clicking a node selects it (the admin page
 * lists that category's products), while the +/- badge on a parent collapses
 * or expands its branch. Node positions come from the fully expanded tree, so
 * toggling a branch never reflows the rest of the map. Layout recomputes
 * whenever `categories` changes, so it stays live after saves.
 */
export default function CategoryTreeMap({
  categories,
  selectedId,
  onSelect,
  emptyLabel,
  emptyNodeLabel,
  itemsLabel,
  hint,
  collapseLabel,
  expandLabel,
  collapseAllLabel,
  expandAllLabel,
}: CategoryTreeMapProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const dragRef = useRef<{
    pointerId: number;
    startX: number;
    startY: number;
    originX: number;
    originY: number;
  } | null>(null);
  // Viewport transform lives in a plain object + React state so wheel and pan
  // handlers always compose against the latest value.
  const [view, setView] = useState<Viewport>({ x: 0, y: 0, scale: 1 });
  // Collapsed branch ids survive data reloads; collapsing only hides nodes, it
  // never changes the taxonomy or the fixed node positions.
  const [collapsedIds, setCollapsedIds] = useState<Set<string>>(() => new Set());

  const toggleCollapse = useCallback((id: string) => {
    setCollapsedIds((previous) => {
      const next = new Set(previous);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const fullLayout = useMemo(() => buildLayout(categories), [categories]);
  const layout = useMemo(() => visibleLayout(fullLayout, collapsedIds), [fullLayout, collapsedIds]);

  const [animatedNodes, setAnimatedNodes] = useState<AnimatedNode[]>(() => toStaticNodes(layout.nodes));
  const animatedNodesRef = useRef(animatedNodes);
  const previousCollapsedRef = useRef(collapsedIds);
  const animationCategoriesRef = useRef(categories);

  // Parent nodes get a collapse toggle and participate in "collapse all".
  const collapsibleIds = useMemo(() => {
    const existing = new Set(categories.map((category) => category.id));
    const ids = new Set<string>();
    for (const category of categories) {
      if (category.parentId && category.parentId !== category.id && existing.has(category.parentId)) {
        ids.add(category.parentId);
      }
    }
    return ids;
  }, [categories]);

  // A parent that is hidden inside a collapsed ancestor is not "expanded" for
  // the user, so base the toggle label on the visible layout instead of the raw
  // collapsed-id set.
  const allCollapsed =
    collapsibleIds.size > 0 &&
    layout.nodes.every((node) => node.childCount === 0 || node.collapsed);

  const toggleAllCollapsed = useCallback(() => {
    setCollapsedIds(allCollapsed ? new Set() : new Set(collapsibleIds));
  }, [allCollapsed, collapsibleIds]);

  const fitToView = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    const width = container.clientWidth;
    const height = container.clientHeight;
    if (width === 0 || height === 0) return;

    const scale = clamp(
      Math.min((width - 48) / Math.max(fullLayout.width, 1), (height - 48) / Math.max(fullLayout.height, 1)),
      MIN_SCALE,
      1,
    );

    setView({
      scale,
      x: (width - fullLayout.width * scale) / 2,
      y: (height - fullLayout.height * scale) / 2,
    });
  }, [fullLayout.width, fullLayout.height]);

  // Initial/re-layout view: keep nodes readable (scale never above 1) and fit
  // the width, centring vertically when the tree is short enough. Tall trees
  // start at the top and are explored by panning; the Fit button shows all.
  const resetView = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    const width = container.clientWidth;
    const height = container.clientHeight;
    if (width === 0 || height === 0) return;

    const scale = clamp(Math.min(1, (width - 48) / Math.max(fullLayout.width, 1)), MIN_SCALE, 1);
    const scaledHeight = fullLayout.height * scale;

    setView({
      scale,
      x: (width - fullLayout.width * scale) / 2,
      y: scaledHeight <= height - 48 ? (height - scaledHeight) / 2 : 24,
    });
  }, [fullLayout.width, fullLayout.height]);

  // Re-frame on first mount and whenever the category data is replaced (after a
  // save or delete). Collapsing a branch must not reset the viewport, otherwise
  // the user loses their pan/zoom position on every toggle.
  const previousCategoriesRef = useRef(categories);
  const hasMountedRef = useRef(false);
  useEffect(() => {
    const categoriesChanged = previousCategoriesRef.current !== categories;
    if (hasMountedRef.current && !categoriesChanged) return;

    hasMountedRef.current = true;
    previousCategoriesRef.current = categories;
    resetView();
  }, [categories, resetView]);

  // Animate only explicit collapse/expand toggles. Data reloads (new category
  // array) snap to the new layout and are re-framed by the effect above.
  useEffect(() => {
    const collapsedChanged = previousCollapsedRef.current !== collapsedIds;
    const categoriesChanged = animationCategoriesRef.current !== categories;
    previousCollapsedRef.current = collapsedIds;
    animationCategoriesRef.current = categories;

    const prefersReducedMotion =
      typeof window !== "undefined" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    const shouldAnimate =
      collapsedChanged &&
      !categoriesChanged &&
      !prefersReducedMotion &&
      animatedNodesRef.current.length > 0 &&
      layout.nodes.length > 0;

    if (!shouldAnimate) {
      const staticNodes = toStaticNodes(layout.nodes);
      animatedNodesRef.current = staticNodes;
      setAnimatedNodes(staticNodes);
      return;
    }

    const steps = buildAnimation(animatedNodesRef.current, layout);
    const start = performance.now();
    let frameId = 0;

    const tick = (now: number) => {
      const progress = Math.min(1, (now - start) / COLLAPSE_DURATION);
      if (progress >= 1) {
        const staticNodes = toStaticNodes(layout.nodes);
        animatedNodesRef.current = staticNodes;
        setAnimatedNodes(staticNodes);
        return;
      }

      const eased = easeInOutCubic(progress);
      const frame = steps.map((step) => ({
        ...step.node,
        displayX: step.fromX + (step.toX - step.fromX) * eased,
        displayY: step.fromY + (step.toY - step.fromY) * eased,
        scale: step.fromScale + (step.toScale - step.fromScale) * eased,
        opacity: step.fromOpacity + (step.toOpacity - step.fromOpacity) * eased,
        entering: step.entering,
        exiting: step.exiting,
      }));
      animatedNodesRef.current = frame;
      setAnimatedNodes(frame);
      frameId = requestAnimationFrame(tick);
    };

    frameId = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frameId);
  }, [layout, collapsedIds, categories]);

  // Native wheel listener: React's synthetic wheel handler is passive, so
  // preventDefault there would not reliably stop the page from scrolling.
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const handleWheel = (event: WheelEvent) => {
      event.preventDefault();

      const rect = container.getBoundingClientRect();
      const pointerX = event.clientX - rect.left;
      const pointerY = event.clientY - rect.top;
      const factor = Math.exp(-event.deltaY * 0.0015);

      setView((current) => {
        const scale = clamp(current.scale * factor, MIN_SCALE, MAX_SCALE);
        const ratio = scale / current.scale;
        return {
          scale,
          x: pointerX - (pointerX - current.x) * ratio,
          y: pointerY - (pointerY - current.y) * ratio,
        };
      });
    };

    container.addEventListener("wheel", handleWheel, { passive: false });
    return () => container.removeEventListener("wheel", handleWheel);
  }, []);

  const zoomBy = (factor: number) => {
    const container = containerRef.current;
    if (!container) return;

    const centerX = container.clientWidth / 2;
    const centerY = container.clientHeight / 2;
    setView((current) => {
      const scale = clamp(current.scale * factor, MIN_SCALE, MAX_SCALE);
      const ratio = scale / current.scale;
      return {
        scale,
        x: centerX - (centerX - current.x) * ratio,
        y: centerY - (centerY - current.y) * ratio,
      };
    });
  };

  const handlePointerDown = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (event.button !== 0) return;
    const target = event.target as Element;
    if (target.closest("[data-map-node]") || target.closest("[data-map-control]")) {
      return;
    }

    event.currentTarget.setPointerCapture(event.pointerId);
    dragRef.current = {
      pointerId: event.pointerId,
      startX: event.clientX,
      startY: event.clientY,
      originX: view.x,
      originY: view.y,
    };
  };

  const handlePointerMove = (event: ReactPointerEvent<HTMLDivElement>) => {
    const drag = dragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;

    setView((current) => ({
      ...current,
      x: drag.originX + (event.clientX - drag.startX),
      y: drag.originY + (event.clientY - drag.startY),
    }));
  };

  const endDrag = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (dragRef.current?.pointerId !== event.pointerId) return;
    dragRef.current = null;
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
  };

  if (fullLayout.nodes.length === 0) {
    return (
      <div className="flex h-[420px] items-center justify-center rounded-2xl border border-dashed border-border bg-secondary/20 text-sm text-muted-foreground">
        {emptyLabel}
      </div>
    );
  }

  const renderEdges = buildEdges(animatedNodes);
  // Paint exiting nodes last so a collapsing branch stays visible while it is
  // pulled into its parent and fades out.
  const orderedNodes = [...animatedNodes].sort(
    (a, b) => Number(a.exiting) - Number(b.exiting),
  );

  return (
    <div
      ref={containerRef}
      onPointerDown={handlePointerDown}
      onPointerMove={handlePointerMove}
      onPointerUp={endDrag}
      onPointerCancel={endDrag}
      className="relative h-[70vh] min-h-[520px] max-h-[820px] w-full touch-none overflow-hidden rounded-2xl border border-border bg-secondary/20 select-none"
    >
      <svg className="h-full w-full" role="img" aria-label={hint}>
        <g transform={`translate(${view.x} ${view.y}) scale(${view.scale})`}>
          {renderEdges.map((edge) => {
            const delta = edge.toX - edge.fromX;
            const curve = delta > 24 ? Math.min(delta / 2, 120) : 0;
            const path =
              curve > 0
                ? `M ${edge.fromX} ${edge.fromY} C ${edge.fromX + curve} ${edge.fromY}, ${edge.toX - curve} ${edge.toY}, ${edge.toX} ${edge.toY}`
                : `M ${edge.fromX} ${edge.fromY} L ${edge.toX} ${edge.toY}`;

            return (
              <path
                key={edge.id}
                d={path}
                className="fill-none stroke-border"
                strokeWidth={1.5}
                style={{ opacity: edge.opacity }}
              />
            );
          })}

          {orderedNodes.map((node) => {
            const isSelected = selectedId === node.id;
            const isRoot = node.depth === 0;
            const count = node.totalProducts;
            const isInteractive = !node.entering && !node.exiting;
            const centerX = NODE_WIDTH / 2;
            const centerY = NODE_HEIGHT / 2;

            return (
              <g
                key={node.id}
                data-map-node
                transform={`translate(${node.displayX} ${node.displayY}) translate(${centerX} ${centerY}) scale(${node.scale}) translate(${-centerX} ${-centerY})`}
                opacity={node.opacity}
                style={{ pointerEvents: isInteractive ? "auto" : "none" }}
                className={isInteractive ? "cursor-pointer outline-none" : "outline-none"}
                role={isInteractive ? "button" : undefined}
                tabIndex={isInteractive ? 0 : -1}
                aria-hidden={isInteractive ? undefined : true}
                aria-label={node.name}
                onClick={isInteractive ? () => onSelect(node) : undefined}
                onKeyDown={
                  isInteractive
                    ? (event) => {
                        if (event.key === "Enter" || event.key === " ") {
                          event.preventDefault();
                          onSelect(node);
                        } else if (event.key === "ArrowRight" && node.childCount > 0) {
                          event.preventDefault();
                          if (node.collapsed) toggleCollapse(node.id);
                        } else if (event.key === "ArrowLeft" && node.childCount > 0) {
                          event.preventDefault();
                          if (!node.collapsed) toggleCollapse(node.id);
                        }
                      }
                    : undefined
                }
              >
                <rect
                  width={NODE_WIDTH}
                  height={NODE_HEIGHT}
                  rx={14}
                  strokeWidth={isSelected ? 2 : 1}
                  className={
                    isSelected
                      ? "fill-accent-brand/15 stroke-accent-brand"
                      : isRoot
                        ? "fill-accent-brand/5 stroke-accent-brand/50"
                        : "fill-card stroke-border"
                  }
                />
                <foreignObject x={0} y={0} width={NODE_WIDTH} height={NODE_HEIGHT}>
                  <div
                    className={`flex h-full w-full flex-col justify-center gap-0.5 text-start ${
                      node.childCount > 0 ? "pl-4 pr-9" : "px-4"
                    }`}
                  >
                    <span className="truncate text-xs font-semibold text-foreground">{node.name}</span>
                    <span className="truncate text-[10px] text-muted-foreground">
                      {count > 0 ? itemsLabel(count) : emptyNodeLabel}
                      {node.collapsed && node.descendantCount > 0 ? ` · +${node.descendantCount}` : ""}
                    </span>
                  </div>
                </foreignObject>
                {isInteractive && node.childCount > 0 && (
                  <g
                    data-map-toggle
                    transform={`translate(${NODE_WIDTH - 16} ${NODE_HEIGHT / 2})`}
                    role="button"
                    aria-label={node.collapsed ? expandLabel : collapseLabel}
                    className="cursor-pointer"
                    onClick={(event) => {
                      event.stopPropagation();
                      toggleCollapse(node.id);
                    }}
                  >
                    <circle r={10} strokeWidth={1} className="fill-background stroke-border" />
                    {node.collapsed ? (
                      <path
                        d="M -3.5 0 H 3.5 M 0 -3.5 V 3.5"
                        strokeWidth={1.8}
                        strokeLinecap="round"
                        className="fill-none stroke-muted-foreground"
                      />
                    ) : (
                      <path
                        d="M -3.5 0 H 3.5"
                        strokeWidth={1.8}
                        strokeLinecap="round"
                        className="fill-none stroke-muted-foreground"
                      />
                    )}
                  </g>
                )}
              </g>
            );
          })}
        </g>
      </svg>

      <div
        data-map-control
        className="absolute end-3 top-3 flex items-center gap-1 rounded-xl border border-border bg-background/90 p-1 shadow-xs backdrop-blur-sm"
      >
        <button
          type="button"
          onClick={() => zoomBy(1 / 1.2)}
          className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
          aria-label="Zoom out"
        >
          <Minus className="size-4" />
        </button>
        <button
          type="button"
          onClick={fitToView}
          className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
          aria-label="Fit to view"
        >
          <Maximize2 className="size-4" />
        </button>
        <button
          type="button"
          onClick={() => zoomBy(1.2)}
          className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
          aria-label="Zoom in"
        >
          <Plus className="size-4" />
        </button>
        {collapsibleIds.size > 0 && (
          <>
            <span className="mx-0.5 h-4 w-px bg-border" aria-hidden />
            <button
              type="button"
              onClick={toggleAllCollapsed}
              className="flex items-center gap-1.5 rounded-lg px-2 py-1.5 text-[11px] font-semibold text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
              aria-label={allCollapsed ? expandAllLabel : collapseAllLabel}
              title={allCollapsed ? expandAllLabel : collapseAllLabel}
            >
              {allCollapsed ? (
                <ChevronsUpDown className="size-4" />
              ) : (
                <ChevronsDownUp className="size-4" />
              )}
              <span className="hidden sm:inline">{allCollapsed ? expandAllLabel : collapseAllLabel}</span>
            </button>
          </>
        )}
      </div>

      <div className="pointer-events-none absolute inset-x-3 bottom-3 text-center text-[10px] text-muted-foreground">
        {hint}
      </div>
    </div>
  );
}
