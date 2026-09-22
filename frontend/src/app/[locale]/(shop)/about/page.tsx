import AboutLayout from "@/features/about/components/layout/aboutLayout";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "About Us | PRIM",
  description: "Learn more about PRIM's values, mission, and team",
};

export default function AboutPage() {
  return <AboutLayout />;
}
