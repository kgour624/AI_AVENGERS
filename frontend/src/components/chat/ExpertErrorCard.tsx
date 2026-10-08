import React from "react";

type Props = {
  title?: string;
  message: string;
  code?: string;
  onRetry?: () => void;
  retryLoading?: boolean;
};

export default function ExpertErrorCard({ title = "Jawab shuru hone se pehle error", message, code, onRetry, retryLoading }: Props) {
  return (
    <div className="rounded-lg border border-red-200 bg-red-50 p-4 shadow-sm" role="alert" data-testid="expert-error-card">
      <div className="flex items-start gap-3">
        <span className="text-red-600 text-xl">⚠️</span>
        <div className="flex-1">
          <div className="font-semibold text-red-800">{title}</div>
          {code && <div className="text-xs text-red-600 mt-0.5">Code: {code}</div>}
          <div className="text-sm text-red-700 mt-1 whitespace-pre-wrap">{message}</div>
          {onRetry && (
            <button
              onClick={onRetry}
              disabled={retryLoading}
              className="mt-3 inline-flex items-center rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-60"
              data-testid="retry-expert-btn"
            >
              {retryLoading ? "Retrying..." : "Retry"}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
