import { Metadata } from "next";
import { ReviewsContent } from "@/features/user";

export const metadata: Metadata = {
  title: "Your Reviews | PRIM",
  description: "View and manage your product reviews",
};

export default function ReviewsPage() {
  return <ReviewsContent />;
}
