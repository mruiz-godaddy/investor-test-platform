interface Props {
  title: string;
  description?: string;
}

export default function EmptyState({ title, description }: Props) {
  return (
    <div className="flex flex-col items-center justify-center rounded-lg border-2 border-dashed border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white p-12 text-center">
      <h3 className="text-sm font-semibold text-gray-900 dark:text-white">{title}</h3>
      {description && <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">{description}</p>}
    </div>
  );
}
