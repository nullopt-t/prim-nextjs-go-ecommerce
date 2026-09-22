import BeliefsGrid from "@/features/about/components/ui/beliefsGrid";
import BeliefsDescription from "@/features/about/components/ui/beliefsDescription";

export default function Beliefs() {
  return (
    <div className="flex flex-col lg:flex-row gap-8 lg:gap-12 justify-between items-start">
      <div className="lg:w-96 shrink-0">
        <BeliefsDescription />
      </div>
      <div className="flex-1">
        <BeliefsGrid />
      </div>
    </div>
  );
}
