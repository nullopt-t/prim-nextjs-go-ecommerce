import AboutHero from "@/modules/content/components/about/ui/hero";
import Beliefs from "@/modules/content/components/about/ui/beliefs";

export default function AboutLayout() {
  return (
    <div className="mx-auto mb-5 flex w-full max-w-5xl flex-col gap-12">
      <AboutHero />
      <Beliefs />
    </div>
  );
}
