import regionSource from './china-region.json'

export interface RegionOption {
  value: string
  label: string
  children?: RegionOption[]
}

export const regionData = regionSource.regionData as RegionOption[]

export const codeToText = regionSource.codeToText as Record<string, string>
