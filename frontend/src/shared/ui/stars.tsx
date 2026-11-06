import { Star } from "lucide-react";

interface StarsProps {
  starsNum?: number | string;
  className?: string;
}

export function Stars({ starsNum = 0, className = "" }: StarsProps) {
  const numericStars = typeof starsNum === "string" ? parseFloat(starsNum) || 0 : starsNum;

  return (
    <span
      className={`flex ${className}`}
      role="img"
      aria-label={`${numericStars} / 5`}
    >
      {Array.from({ length: 5 }).map((_, index) => {
        const isFilled = index < numericStars;

        return (
          <Star
            key={index}
            aria-hidden
            className={`size-4 ${
              isFilled ? "text-yellow-500" : "text-gray-400"
            }`}
            fill={isFilled ? "currentColor" : "none"}
            stroke="currentColor"
          />
        );
      })}
    </span>
  );
}
