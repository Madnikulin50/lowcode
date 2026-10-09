import moment, { MomentInput } from 'moment'
import { currentLanguage } from './locale'

declare type DateTimeInput = unknown | MomentInput

declare type DateTimeFormatOptions = Intl.DateTimeFormatOptions & {
  dateStyle?: 'full' | 'long' | 'medium' | 'short';
  timeStyle?: 'full' | 'long' | 'medium' | 'short';
}

/**
 * Parses input into Date using Moment library
 *
 * @param input
 */
function parse (input: DateTimeInput): Date {
  return moment(input as MomentInput).toDate()
}

function format (input: DateTimeInput, options: DateTimeFormatOptions): string {
  return (new Intl.DateTimeFormat(currentLanguage(), options)).format(parse(input))
}

/**
 * Outputs locally formatted date and time, no seconds
 *
 * Examples:
 * "September 8, 2021 at 9:41 AM"
 * "08 september 2021 09:41"
 * "8.09.2021 09:42"
 *
 * @param input
 * @param options
 */
export function fullDateTime (input: DateTimeInput, options: DateTimeFormatOptions = { dateStyle: 'medium', timeStyle: 'short' }): string {
  return format(input, options)
}

/**
 * Outputs locally formatted date without time
 *
 * Example:
 * 09/04/1986
 *
 * @param input
 * @param options
 */
export function date (input: DateTimeInput, options: DateTimeFormatOptions = { dateStyle: 'short' }): string {
  return format(input, options)
}

/**
 * Outputs locally formatted time
 *
 * Example:
 * 8:30 PM
 *
 * @param input
 * @param options
 */
export function time (input: DateTimeInput, options: DateTimeFormatOptions = { timeStyle: 'short' }): string {
  return format(input, options)
}
