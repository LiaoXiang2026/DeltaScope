import type { ReactNode } from "react";

interface SectionProps {
  title: string;
  description?: string;
  children: ReactNode;
}

export function WorkbenchSection({ title, description, children }: SectionProps) {
  return (
    <section className="surface-card panel-card">
      <div className="panel-header">
        <h2>{title}</h2>
        {description ? <p>{description}</p> : null}
      </div>
      <div className="panel-body">{children}</div>
    </section>
  );
}

interface FieldProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  placeholder?: string;
  helper?: string;
  actionLabel?: string;
  onAction?: () => void;
}

export function WorkbenchField({
  label,
  value,
  onChange,
  type = "text",
  placeholder,
  helper,
  actionLabel,
  onAction,
}: FieldProps) {
  return (
    <label className="field-stack">
      <span className="field-label">{label}</span>
      <div className="field-row">
        <input
          type={type}
          value={value}
          placeholder={placeholder}
          onChange={(event) => onChange(event.target.value)}
          className="field-control"
        />
        {actionLabel && onAction ? (
          <button type="button" onClick={onAction} className="field-button">
            {actionLabel}
          </button>
        ) : null}
      </div>
      {helper ? <span className="field-helper">{helper}</span> : null}
    </label>
  );
}

interface ToggleProps {
  label: string;
  description: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
}

export function WorkbenchToggle({ label, description, checked, onChange }: ToggleProps) {
  return (
    <label className="toggle-card">
      <div className="toggle-copy-block">
        <p className="toggle-label">{label}</p>
        <p className="toggle-description">{description}</p>
      </div>
      <span className="toggle-switch">
        <input type="checkbox" checked={checked} onChange={(event) => onChange(event.target.checked)} />
        <span className="toggle-switch-track" />
      </span>
    </label>
  );
}

interface StatusPillProps {
  tone?: "neutral" | "ready" | "running" | "warning";
  children: ReactNode;
}

export function StatusPill({ tone = "neutral", children }: StatusPillProps) {
  return <span className={`status-pill status-pill--${tone}`}>{children}</span>;
}
