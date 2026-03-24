"use client";

export function AiThinking({ text }: { text: string }) {
  return (
    <span className="ai-thinking" role="status" aria-live="polite">
      {text}
      <span className="ai-dots" aria-hidden="true">
        <span />
        <span />
        <span />
      </span>
    </span>
  );
}
