import { Code, ConnectError } from "@connectrpc/connect";

function looksSensitive(msg: string): boolean {
  const s = msg.toLowerCase();
  return (
    /\bsql\b/.test(s) ||
    /\bsql:\s*/.test(s) ||
    /\bselect\b|\binsert\b|\bupdate\b|\bdelete\b/.test(s) ||
    /\bfrom\s+[`"\w.]+\b/.test(s) ||
    /\binto\s+[`"\w.]+\b/.test(s) ||
    /\bupdate\s+[`"\w.]+\b/.test(s) ||
    /\bdelete\s+from\s+[`"\w.]+\b/.test(s) ||
    /duplicate entry|foreign key|constraint|syntax error/.test(s) ||
    /unknown column|table .* doesn't exist|access denied/.test(s) ||
    /[`"].+[`"]/.test(s)
  );
}

export function toAdminErrorMessage(err: unknown, fallback: string): string {
  if (err instanceof ConnectError) {
    if (err.code === Code.Unauthenticated || err.code === Code.PermissionDenied) {
      return "認証エラーです。一度ログアウトして再ログインするか、管理者キーで入り直してください。";
    }

    if (err.code === Code.InvalidArgument) {
      const msg = err.rawMessage ?? "";
      return msg && msg.length <= 200 && !looksSensitive(msg) ? msg : fallback;
    }

    if (err.code === Code.FailedPrecondition) {
      const msg = err.rawMessage ?? "";
      return msg && msg.length <= 200 && !looksSensitive(msg) ? msg : fallback;
    }

    return fallback;
  }

  return fallback;
}

