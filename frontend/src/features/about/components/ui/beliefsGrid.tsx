"use client";

import { Star, Gauge, ShieldCheck, Smile } from "lucide-react";
import BeliefCards, { BeliefItem } from "@/features/about/components/ui/beliefsCards";

export default function BeliefsGrid() {
  const beliefsArr: BeliefItem[] = [
    {
      id: "quality-cdkt",
      title: "values.quality.title",
      subTitle: "values.quality.description",
      icon: Star,
    },
    {
      id: "speed-cdkt",
      title: "values.speed.title",
      subTitle: "values.speed.description",
      icon: Gauge,
    },
    {
      id: "trust-cdkt",
      title: "values.trust.title",
      subTitle: "values.trust.description",
      icon: ShieldCheck,
    },
    {
      id: "Simplicity-cdkt",
      title: "values.simplicity.title",
      subTitle: "values.simplicity.description",
      icon: Smile,
    },
  ];

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
      {beliefsArr.map((belief) => (
        <BeliefCards key={belief.id} belief={belief} />
      ))}
    </div>
  );
}
