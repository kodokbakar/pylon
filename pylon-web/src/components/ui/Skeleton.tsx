import type { HTMLAttributes } from 'react'

type SkeletonProps = HTMLAttributes<HTMLDivElement>

export function Skeleton({ className = '', ...props }: SkeletonProps) {
  return (
    <div
      aria-hidden="true"
      className={[
        'animate-pulse border border-[var(--color-line)] bg-[var(--color-grid)]',
        className,
      ]
        .filter(Boolean)
        .join(' ')}
      {...props}
    />
  )
}
