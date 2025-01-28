import * as React from "react"

const Card = React.forwardRef<
    HTMLDivElement,
    React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
    <div
        ref={ref}
        className={`rounded-lg border border-input-border bg-card-background text-card-foreground shadow-sm ${className}`}
        style={{
            backgroundColor: 'var(--card-background)',
            color: 'var(--card-foreground)',
            borderColor: 'var(--input-border)'
        }}
        {...props}
    />
))
Card.displayName = "Card"

const CardHeader = React.forwardRef<
    HTMLDivElement,
    React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
    <div
        ref={ref}
        className={`flex flex-col space-y-1.5 p-6 ${className}`}
        {...props}
    />
))
CardHeader.displayName = "CardHeader"

const CardTitle = React.forwardRef<
    HTMLParagraphElement,
    React.HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => (
    <h3
        ref={ref}
        className={`text-lg font-semibold leading-none tracking-tight ${className}`}
        style={{ color: 'var(--card-foreground)' }}
        {...props}
    />
))
CardTitle.displayName = "CardTitle"

const CardContent = React.forwardRef<
    HTMLDivElement,
    React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
    <div
        ref={ref}
        className={`p-6 pt-0 ${className}`}
        {...props}
    />
))
CardContent.displayName = "CardContent"

export { Card, CardHeader, CardTitle, CardContent }