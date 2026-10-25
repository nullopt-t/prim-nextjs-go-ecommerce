import React from "react";

export default function PaymentMethods() {
  const methods = ["Visa", "MC", "PayPal", "Apple Pay"];

  return (
    <div className="flex justify-center gap-2.5 text-muted-foreground text-txt-sm">
      {methods.map((method) => (
        <span
          key={method}
          className="bg-secondary text-secondary-foreground p-1 px-2.5 rounded-md text-xs font-medium"
        >
          {method}
        </span>
      ))}
    </div>
  );
}
