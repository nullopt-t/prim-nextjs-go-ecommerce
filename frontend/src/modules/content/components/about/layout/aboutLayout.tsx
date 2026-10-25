import AboutHero from "@/modules/content/components/about/ui/hero";
import Beliefs from "@/modules/content/components/about/ui/beliefs";

export default function AboutLayout() {
  return (
    <div className="mb-5 flex flex-col gap-12">
      <AboutHero />
      <Beliefs />
    </div>
  );
}
