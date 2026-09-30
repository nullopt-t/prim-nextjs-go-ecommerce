import { Truck, ShieldCheck } from "lucide-react";

interface ProductSpecsProps {
  productName: string;
  category: string;
  brand?: string;
  description?: string;
  attributes?: Record<string, any>;
}

export function ProductSpecs({ productName, category, brand, description, attributes }: ProductSpecsProps) {
  const displayBrand = brand || productName.split(" ")[0] || "PRIM";

  return (
    <div className="mt-20 flex flex-col gap-16">
      {/* Description Section */}
      <section>
        <h2 className="text-2xl font-bold text-foreground mb-6">Description</h2>
        <div className="prose prose-sm max-w-none text-muted-foreground bg-card border border-border rounded-2xl p-6 md:p-8 shadow-sm">
          {description ? (
            <p className="whitespace-pre-line leading-relaxed text-foreground/90">{description}</p>
          ) : (
            <>
              <p>
                Enhance your lifestyle with the {productName}. Built with premium materials 
                and cutting-edge technology, it delivers outstanding performance and reliability. 
                Whether you're a professional or enthusiast, you'll appreciate the attention to 
                detail and superior build quality.
              </p>
              <ul className="mt-4 space-y-2">
                <li>High-quality materials for lasting durability</li>
                <li>Ergonomic design for maximum comfort</li>
                <li>Advanced features for enhanced productivity</li>
                <li>Eco-friendly packaging and sustainable production</li>
              </ul>
            </>
          )}
        </div>
      </section>

      {/* Specifications Section */}
      <section>
        <h2 className="text-2xl font-bold text-foreground mb-6">Specifications</h2>
        <div className="bg-card border border-border rounded-2xl p-6 md:p-8 shadow-sm">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-12 gap-y-4 text-sm">
            <div className="flex justify-between py-3 border-b border-border/50">
              <span className="text-muted-foreground">Brand</span>
              <span className="font-medium text-foreground">{displayBrand}</span>
            </div>
            <div className="flex justify-between py-3 border-b border-border/50">
              <span className="text-muted-foreground">Category</span>
              <span className="font-medium text-foreground capitalize">{category.replace("-", " ")}</span>
            </div>
            {attributes &&
              Object.entries(attributes).map(([key, val]) => (
                <div key={key} className="flex justify-between py-3 border-b border-border/50">
                  <span className="text-muted-foreground capitalize">{key}</span>
                  <span className="font-medium text-foreground">{String(val)}</span>
                </div>
              ))}
            <div className="flex justify-between py-3 border-b border-border/50">
              <span className="text-muted-foreground">Condition</span>
              <span className="font-medium text-foreground">Brand New</span>
            </div>
            <div className="flex justify-between py-3 border-b border-border/50">
              <span className="text-muted-foreground">Warranty</span>
              <span className="font-medium text-foreground">1 Year Manufacturer</span>
            </div>
          </div>
        </div>
      </section>

      {/* Shipping & Returns Section */}
      <section>
        <h2 className="text-2xl font-bold text-foreground mb-6">Shipping & Returns</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-card border border-border rounded-2xl p-6 flex flex-col gap-3 shadow-sm hover:border-accent-brand/50 transition-colors">
            <div className="w-10 h-10 rounded-full bg-secondary/50 flex items-center justify-center text-foreground">
              <Truck className="size-5" />
            </div>
            <h4 className="text-foreground font-bold text-base">Shipping Information</h4>
            <p className="text-sm text-muted-foreground leading-relaxed">
              We offer free standard shipping on all orders worldwide. Delivery typically takes 3-5 business days depending on your location. Expedited shipping is available at checkout for an additional fee.
            </p>
          </div>

          <div className="bg-card border border-border rounded-2xl p-6 flex flex-col gap-3 shadow-sm hover:border-accent-brand/50 transition-colors">
            <div className="w-10 h-10 rounded-full bg-secondary/50 flex items-center justify-center text-foreground">
              <ShieldCheck className="size-5" />
            </div>
            <h4 className="text-foreground font-bold text-base">Return Policy</h4>
            <p className="text-sm text-muted-foreground leading-relaxed">
              If you're not completely satisfied with your purchase, you can return it within 30 days for a full refund or exchange. The item must be in its original condition and packaging. Please contact our support team to initiate a return.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}
