interface PaginationProps {
  page?: number;
  total?: number;
  onNavigate: (page: number) => void;
  class?: string;
}

const BOUNDARY = 2;

export function Pagination({ page = 1, total = 1, onNavigate, class: className }: PaginationProps) {
  if (total <= 1) return null;

  const count = Math.min(total, BOUNDARY * 2 + 1);
  const start = Math.max(1, Math.min(page - BOUNDARY, total - count + 1));
  const pages = Array.from({ length: count }, (_, index) => start + index);

  return (
    <nav class={`join ${className}`} aria-label="Pagination">
      <button
        type="button"
        class="join-item btn"
        aria-label="Previous page"
        disabled={page <= 1}
        onClick={() => onNavigate(page - 1)}
      >
        <span class="icon-[mdi--chevron-left]"></span>
      </button>
      {pages.map((value) => (
        <button
          key={value}
          type="button"
          class={`join-item btn${value === page ? " btn-active" : ""}`}
          aria-label={`Page ${value}`}
          aria-current={value === page ? "page" : undefined}
          onClick={value !== page ? () => onNavigate(value) : undefined}
        >
          {value}
        </button>
      ))}
      <button
        type="button"
        class="join-item btn"
        aria-label="Next page"
        disabled={page >= total}
        onClick={() => onNavigate(page + 1)}
      >
        <span class="icon-[mdi--chevron-right]"></span>
      </button>
    </nav>
  );
}
