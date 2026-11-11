/**
 * Shared serialization of the storefront review query params for the backend
 * `/api/v1/reviews` contract. Used by the server fetcher and the client
 * catalog service so the two can never drift apart:
 * - the page size is fixed so both sides ask for the same slice;
 * - the star filter is applied by the backend, so the listed reviews always
 *   match the histogram the backend counted;
 * - paging is cursor-based: the caller sends back the opaque cursor the
 *   previous response returned.
 */
export interface ReviewQuery {
	/** Cursor for the next page, as returned by the previous response. */
	cursor?: string | null;
	/** Reviews per page. */
	pageSize?: number;
	/** Restricts the listing to one star score (1-5). */
	rating?: number | null;
}

/** Reviews fetched per page on the product detail route. */
export const REVIEWS_PAGE_SIZE = 20;

export function buildReviewQuery(query: ReviewQuery = {}): URLSearchParams {
	const params = new URLSearchParams();
	if (query.cursor) params.set("cursor", query.cursor);
	if (query.pageSize) params.set("pageSize", String(query.pageSize));
	if (query.rating) params.set("rating", String(query.rating));
	return params;
}
