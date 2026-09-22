import React from "react";

interface QuantitySelectorProps {
  value?: number;
  initialValue?: number;
  min?: number;
  max?: number;
  disabled?: boolean;
  onChange?: (quantity: number) => void;
}

export function QuantitySelector({
  value = 1,
  min = 1,
  max = 99,
  onChange,
}: QuantitySelectorProps) {
  const handleDecrement = () => {
    if (value > min) {
      onChange?.(value - 1);
    }
  };

  const handleIncrement = () => {
    if (value < max) {
      onChange?.(value + 1);
    }
  };

  return (
    <div className="text-foreground flex items-center border border-border rounded-md w-24 md:w-36 text-txt-sm md:text-txt-md lg:text-txt-lg select-none">
      <button
        type="button"
        onClick={handleDecrement}
        disabled={value <= min}
        className="py-1 flex-1 text-center border-r border-border hover:bg-accent hover:text-accent-foreground disabled:opacity-40"
      >
        -
      </button>
      <span className="py-1 flex-[2] text-center">{value}</span>
      <button
        type="button"
        onClick={handleIncrement}
        disabled={value >= max}
        className="py-1 flex-1 text-center border-l border-border hover:bg-accent hover:text-accent-foreground disabled:opacity-40"
      >
        +
      </button>
    </div>
  );
}

export default QuantitySelector;
