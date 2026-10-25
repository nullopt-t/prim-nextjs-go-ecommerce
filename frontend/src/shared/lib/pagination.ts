/**
 * Shared types for keyset (cursor) pagination.
 *
 * List endpoints return a forward cursor only; clients keep the chain of
 * cursors walked so far to serve "previous" without a reverse query.
 */
export interface CursorMeta {
  /** Opaque token for the following page; empty when the list is exhausted. */
  nextCursor: string;
  /** Whether at least one more row exists after the current page. */
  hasMore: boolean;
  /** Maximum number of rows per page. */
  pageSize: number;
}

export const emptyCursorMeta: CursorMeta = {
  nextCursor: "",
  hasMore: false,
  pageSize: 0,
};
