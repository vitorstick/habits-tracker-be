# API Validation Rules

This document lists all validation rules for the Habit Tracker API. Use these rules on the frontend for immediate feedback and on the backend for security.

## CreateHabit Request

### Title
- **Required**: Yes
- **Type**: String
- **Min Length**: 1 character
- **Max Length**: 100 characters
- **Error Messages**:
  - Empty: `"title is required"`
  - Too long: `"title too long (max 100 characters)"`

**Examples:**
```javascript
// ✅ Valid
"Drink Water"
"Exercise for 30 minutes"
"Read before bed"

// ❌ Invalid
""                    // Empty
"a".repeat(101)       // Too long
```

### Color
- **Required**: No (defaults to `#58cc02`)
- **Type**: String
- **Format**: Hex color code
- **Allowed Formats**:
  - `#RGB` (3-digit hex, e.g., `#F00`)
  - `#RRGGBB` (6-digit hex, e.g., `#FF0000`)
- **Case**: Insensitive (both uppercase and lowercase allowed)
- **Error Message**: `"invalid color format (use #RRGGBB or #RGB)"`

**Validation Regex:**
```javascript
const colorRegex = /^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$/;
```

**Examples:**
```javascript
// ✅ Valid
"#FF0000"
"#ff0000"
"#F00"
"#f00"
"#58cc02"
""              // Will use default

// ❌ Invalid
"red"           // Not hex
"FF0000"        // Missing #
"#FF00"         // Wrong length (4 chars)
"#GGGGGG"       // Invalid hex characters
"rgb(255,0,0)"  // Wrong format
```

### Description
- **Required**: No
- **Type**: String
- **Max Length**: None (but consider UX limits on frontend)
- **Stored as**: NULL if empty, string if provided

**Examples:**
```javascript
// ✅ Valid
"Track daily water intake"
""              // Will be stored as NULL
```

### Frequency
- **Required**: No (defaults to `"daily"`)
- **Type**: String
- **Allowed Values**:
  - `"daily"`
  - `"weekly"`
  - `"monthly"`
  - `"custom"`
- **Case**: Sensitive (must be lowercase)
- **Error Message**: `"invalid frequency (must be daily, weekly, monthly, or custom)"`

**Examples:**
```javascript
// ✅ Valid
"daily"
"weekly"
"monthly"
"custom"
""              // Will use default

// ❌ Invalid
"Daily"         // Wrong case
"hourly"        // Not in allowed list
"bi-weekly"     // Not in allowed list
```

### Icon
- **Required**: No
- **Type**: String
- **Format**: Usually emoji or icon identifier
- **Max Length**: None (but consider reasonable limits)

**Examples:**
```javascript
// ✅ Valid
"💧"
"🏃"
"📚"
""              // Empty is allowed
```

### FrequencyDetails
- **Required**: No
- **Type**: JSON Object
- **Format**: Must be valid JSON
- **Structure**: Flexible (depends on frequency type)
- **Error Message**: `"invalid frequencyDetails JSON: [error]"`

**Weekly Format:**
```javascript
{
  "days": [0, 1, 2, 3, 4, 5, 6]  // 0=Sunday, 6=Saturday
}
// OR
{
  "days": ["monday", "wednesday", "friday"]  // Case-insensitive
}
```

**Monthly Format:**
```javascript
{
  "dayOfMonth": 15  // 1-31
}
```

**Examples:**
```javascript
// ✅ Valid
{ "days": [0, 1, 2] }
{ "days": ["monday", "wednesday"] }
{ "dayOfMonth": 15 }
null            // No details
undefined       // No details

// ❌ Invalid
"not json"      // Not valid JSON
{ days: [1, 2] }  // Unquoted keys (if sent as string)
```

### Locked
- **Required**: No (defaults to `false`)
- **Type**: Boolean
- **Allowed Values**: `true` or `false`

**Examples:**
```javascript
// ✅ Valid
true
false
undefined       // Will use default false
```

---

## Complete Request Examples

### Minimal Valid Request
```json
{
  "title": "Drink Water"
}
```

### Full Valid Request
```json
{
  "title": "Exercise",
  "description": "30 minutes of cardio",
  "icon": "🏃",
  "color": "#FF5722",
  "frequency": "weekly",
  "frequencyDetails": {
    "days": ["monday", "wednesday", "friday"]
  },
  "locked": false
}
```

### Invalid Requests with Expected Errors

```json
// Empty title
{
  "title": ""
}
// Error: "title is required"

// Invalid color
{
  "title": "Test",
  "color": "red"
}
// Error: "invalid color format (use #RRGGBB or #RGB)"

// Invalid frequency
{
  "title": "Test",
  "frequency": "hourly"
}
// Error: "invalid frequency (must be daily, weekly, monthly, or custom)"

// Invalid frequencyDetails JSON
{
  "title": "Test",
  "frequencyDetails": "not json"
}
// Error: "invalid frequencyDetails JSON: [parse error]"
```

---

## Frontend Validation Helper (JavaScript/TypeScript)

```javascript
// Validation helpers for frontend
export const validators = {
  title: (value) => {
    if (!value || value.trim() === '') {
      return 'Title is required';
    }
    if (value.length > 100) {
      return 'Title too long (max 100 characters)';
    }
    return null;
  },

  color: (value) => {
    if (!value) return null; // Optional field
    const regex = /^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$/;
    if (!regex.test(value)) {
      return 'Invalid color format (use #RRGGBB or #RGB)';
    }
    return null;
  },

  frequency: (value) => {
    if (!value) return null; // Optional field
    const validFrequencies = ['daily', 'weekly', 'monthly', 'custom'];
    if (!validFrequencies.includes(value)) {
      return 'Invalid frequency (must be daily, weekly, monthly, or custom)';
    }
    return null;
  },

  frequencyDetails: (value) => {
    if (!value) return null; // Optional field
    try {
      JSON.parse(JSON.stringify(value)); // Validate it's valid JSON
      return null;
    } catch (e) {
      return `Invalid frequencyDetails JSON: ${e.message}`;
    }
  }
};

// Usage example
function validateCreateHabitRequest(data) {
  const errors = {};

  const titleError = validators.title(data.title);
  if (titleError) errors.title = titleError;

  const colorError = validators.color(data.color);
  if (colorError) errors.color = colorError;

  const frequencyError = validators.frequency(data.frequency);
  if (frequencyError) errors.frequency = frequencyError;

  const detailsError = validators.frequencyDetails(data.frequencyDetails);
  if (detailsError) errors.frequencyDetails = detailsError;

  return Object.keys(errors).length > 0 ? errors : null;
}
```

---

## Query Parameters

### GET /api/habits

#### date
- **Required**: No (defaults to today)
- **Type**: String
- **Format**: `YYYY-MM-DD` (ISO 8601 date)
- **Error Message**: `"Invalid date format (YYYY-MM-DD)"`

**Examples:**
```javascript
// ✅ Valid
"2024-01-15"
"2024-12-31"
undefined       // Uses today's date

// ❌ Invalid
"15-01-2024"    // Wrong format
"2024/01/15"    // Wrong separator
"invalid"       // Not a date
```

### POST /api/habits/{id}/log

#### date
- **Required**: No (defaults to today)
- **Type**: String
- **Format**: `YYYY-MM-DD` (ISO 8601 date)
- **Error Message**: `"Invalid date (use YYYY-MM-DD)"`

#### id (URL parameter)
- **Required**: Yes
- **Type**: Integer
- **Min Value**: 1
- **Error Message**: `"Invalid habit id"`

**Examples:**
```javascript
// ✅ Valid
/api/habits/1/log
/api/habits/42/log?date=2024-01-15

// ❌ Invalid
/api/habits/abc/log       // Not a number
/api/habits/0/log         // Zero not allowed
/api/habits/-1/log        // Negative not allowed
```

---

## Frontend Validation Strategy

### When to Validate

1. **On Input Change** (Real-time):
   - Title length
   - Color format
   - Basic format checks

2. **On Field Blur**:
   - Required fields
   - Complete validation

3. **On Form Submit**:
   - Full validation of all fields
   - Show all errors at once

### User Experience Tips

1. **Show errors below the field** with red text
2. **Use green checkmarks** for valid fields
3. **Disable submit button** until all required fields are valid
4. **Provide autocomplete** for frequency (dropdown with 4 options)
5. **Color picker** should ensure valid hex format
6. **Show character count** for title (e.g., "45/100")
7. **Date picker** should format as YYYY-MM-DD automatically

---

## Testing Checklist

Use this checklist to ensure your frontend validation matches the backend:

- [ ] Empty title rejected
- [ ] Title with 101 characters rejected
- [ ] Title with 100 characters accepted
- [ ] Valid hex colors (#RGB and #RRGGBB) accepted
- [ ] Invalid colors (no #, wrong length, invalid chars) rejected
- [ ] All 4 frequency values accepted
- [ ] Invalid frequency values rejected
- [ ] Valid JSON in frequencyDetails accepted
- [ ] Invalid JSON in frequencyDetails rejected
- [ ] Date format YYYY-MM-DD accepted
- [ ] Invalid date formats rejected
- [ ] Optional fields can be omitted
- [ ] Default values applied correctly
