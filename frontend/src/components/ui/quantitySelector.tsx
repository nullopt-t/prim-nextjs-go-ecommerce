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
  value: controlledValue,
  initialValue = 1,
  min = 1,
  max = 99,
  disabled = false,
  onChange,
}: QuantitySelectorProps) {
  const [internalValue, setInternalValue] = React.useState(
    controlledValue !== undefined ? controlledValue : initialValue
  );

  React.useEffect(() => {
    if (controlledValue !== undefined) {
      setInternalValue(controlledValue);
    }
  }, [controlledValue]);

  const currentValue = controlledValue !== undefined ? controlledValue : internalValue;

  const handleDecrement = () => {
    if (disabled) return;
    if (currentValue > min) {
      const next = currentValue - 1;
      setInternalValue(next);
      onChange?.(next);
    }
  };

  const handleIncrement = () => {
    if (disabled) return;
    if (currentValue < max) {
      const next = currentValue + 1;
      setInternalValue(next);
      onChange?.(next);
    }
  };

  return (
    <div className={`text-foreground flex items-center border border-border/80 rounded-xl h-full w-full select-none ${disabled ? "opacity-50 cursor-not-allowed bg-muted/20" : ""}`}>
      <button
        type="button"
        onClick={handleDecrement}
        disabled={disabled || currentValue <= min}
        className="h-full flex-1 flex items-center justify-center border-r border-border/80 hover:bg-secondary/70 transition-colors disabled:opacity-30 disabled:pointer-events-none cursor-pointer text-sm font-semibold"
      >
        -
      </button>
      <span className="h-full flex-[1.5] flex items-center justify-center font-semibold text-xs sm:text-sm">
        {currentValue}
      </span>
      <button
        type="button"
        onClick={handleIncrement}
        disabled={disabled || currentValue >= max}
        className="h-full flex-1 flex items-center justify-center border-l border-border/80 hover:bg-secondary/70 transition-colors disabled:opacity-30 disabled:pointer-events-none cursor-pointer text-sm font-semibold"
      >
        +
      </button>
    </div>
  );
}

export default QuantitySelector;
