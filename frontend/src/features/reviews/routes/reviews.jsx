import ReviewsLayout from "@/features/reviews/components/layout/reviewsLayout";

export function Reviews() {
	const reviews = [
		{
			id: 1,
			productName:
				"oraimo Watch Nova 2N 1.93'' AMOLED 33-Day Standby Smart Watch",
			starsNumber: 5,
			review: "Stay connected, active, and in control with this stylish smartwatch. The display is bright and clear, the battery lasts for a long time, and the fitness tracking features are very useful. Overall, it is a great smartwatch for everyday use.",
			commentedAt: "2026-07-20",
		},
		{
			id: 2,
			productName: "Anker Soundcore Life Q30 Wireless Headphones",
			starsNumber: 4,
			review: "The sound quality is excellent, especially for the price. The noise cancellation works well, and the headphones are comfortable enough to wear for several hours. The only downside is that the microphone quality could be better.",
			commentedAt: "2026-07-18",
		},
		{
			id: 3,
			productName: "Logitech MX Master 3S Wireless Mouse",
			starsNumber: 5,
			review: "This mouse feels incredibly comfortable and smooth to use. The scrolling is fast and precise, and the battery lasts for weeks. It is definitely worth it if you spend a lot of time working on a computer.",
			commentedAt: "2026-07-15",
		},
		{
			id: 4,
			productName: "Apple AirPods Pro 2nd Generation",
			starsNumber: 4,
			review: "The sound quality is impressive, and the noise cancellation is one of the best I have tried. The case is compact and the battery life is good. I only wish the earbuds were a little more comfortable for long listening sessions.",
			commentedAt: "2026-07-12",
		},
		{
			id: 5,
			productName: "Samsung Galaxy Buds3 Pro",
			starsNumber: 5,
			review: "These earbuds have great sound quality and a very comfortable design. The connection is stable, and the touch controls are easy to use. I am very happy with the purchase and would definitely recommend them.",
			commentedAt: "2026-07-08",
		},
	];

	return <ReviewsLayout reviews={reviews} />;
}
