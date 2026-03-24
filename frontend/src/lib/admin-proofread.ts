export function buildProofreadText(
  title: string,
  bodyMarkdown: string,
  summary: string
): string {
  const parts: string[] = [];
  const t = title.trim();
  const b = bodyMarkdown.trim();
  const s = summary.trim();
  if (t) parts.push(`# ${t}`);
  if (b) parts.push(b);
  if (s) parts.push(`【要約】\n${s}`);
  return parts.join("\n\n");
}
