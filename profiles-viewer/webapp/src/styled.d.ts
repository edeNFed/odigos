import 'styled-components';

declare module 'styled-components' {
  export interface DefaultTheme {
    colors: {
      primary: string;
      background: string;
      surface: string;
      surfaceLight: string;
      text: string;
      textSecondary: string;
      border: string;
      error: string;
      success: string;
    };
  }
}
