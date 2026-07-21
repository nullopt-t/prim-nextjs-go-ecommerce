import { Star } from "lucide-react";

export default function Stars({ starsNum = 0 }) {
	return (
		<div className="flex">
			{Array.from({ length: 5 }).map((_, index) => {
				const isFilled = index < starsNum;

				return (
					<Star
						key={index}
						className={`size-4 ${
							isFilled ? "text-yellow-500" : "text-gray-400"
						}`}
						fill={isFilled ? "currentColor" : "none"}
						stroke="currentColor"
					/>
				);
			})}
		</div>
	);
}
