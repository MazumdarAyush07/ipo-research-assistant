"use client";

import { useState, useRef, useEffect } from "react";
import { ChevronDown, Check } from "lucide-react";

interface Option {
  value: string;
  label: string;
}

interface CustomSelectProps {
  value: string;
  onChange: (val: string) => void;
  options: Option[];
}

export function CustomSelect({ value, onChange, options }: CustomSelectProps) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const selectedOption = options.find((opt) => opt.value === value) || options[0];

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div className="relative" ref={containerRef}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between w-40 min-w-[140px] px-3 py-2 text-sm text-left text-white bg-[#1a1b23] border border-white/10 rounded-lg hover:border-indigo-500/50 transition-colors focus:outline-none focus:ring-1 focus:ring-indigo-500"
      >
        <span className="truncate">{selectedOption.label}</span>
        <ChevronDown className={`w-4 h-4 text-gray-400 transition-transform ${isOpen ? "rotate-180" : ""}`} />
      </button>

      {isOpen && (
        <div className="absolute z-50 w-full min-w-[180px] py-1 mt-1 overflow-auto text-sm bg-[#1a1b23] border border-white/10 rounded-lg shadow-xl max-h-60">
          {options.map((option) => (
            <button
              key={option.value}
              type="button"
              className={`flex items-center justify-between w-full px-3 py-2 text-left hover:bg-white/5 transition-colors ${
                value === option.value ? "text-indigo-400 font-medium" : "text-gray-300"
              }`}
              onClick={() => {
                onChange(option.value);
                setIsOpen(false);
              }}
            >
              <span className="truncate">{option.label}</span>
              {value === option.value && <Check className="w-4 h-4 ml-2" />}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
