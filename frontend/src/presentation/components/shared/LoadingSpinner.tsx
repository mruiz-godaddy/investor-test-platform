export default function LoadingSpinner() {
  return (
    <div className="flex items-center justify-center p-8">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-300 dark:border-gray-600 dark:bg-gray-800 dark:text-white border-t-indigo-600" />
    </div>
  );
}
