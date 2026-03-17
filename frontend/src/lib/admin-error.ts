import { Code, ConnectError } from "@connectrpc/connect";

function looksSensitive(msg: string): boolean {
  const s = msg.toLowerCase();
  // SQL や DB 内部の表現が混ざりやすいパターン（テーブル名そのものではなく「文脈」で検知）
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
    // 認証/権限系は分かる文言を返す（内部情報ではない）
    if (err.code === Code.Unauthenticated || err.code === Code.PermissionDenied) {
      return "認証エラーです。一度ログアウトして再ログインするか、管理者キーで入り直してください。";
    }

    // 入力系は基本そのまま。ただし内部情報っぽいものは隠す
    if (err.code === Code.InvalidArgument) {
      const msg = err.rawMessage ?? "";
      return msg && msg.length <= 200 && !looksSensitive(msg) ? msg : fallback;
    }

    // FailedPrecondition は「AI が利用できません」など運用上の前提不足を示すので表示してよい
    if (err.code === Code.FailedPrecondition) {
      const msg = err.rawMessage ?? "";
      return msg && msg.length <= 200 && !looksSensitive(msg) ? msg : fallback;
    }

    // それ以外（Internal など）は内部情報を隠す
    return fallback;
  }

  // ConnectError 以外（fetch/ネットワーク/例外など）は内部情報が混ざる可能性があるため原則 fallback に倒す
  return fallback;
}

