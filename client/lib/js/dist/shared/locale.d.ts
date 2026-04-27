/**
 * Locale-related helper functions
 */
/**
 * Get the first day of the week for a given locale.
 * Returns 0 for Sunday, 1 for Monday, etc.
 *
 * @param locale - BCP 47 language tag (e.g., 'en-US', 'hu-HU')
 * @returns The first day of the week (0-6)
 */
export declare function getWeekStartDay(locale: string): number;
