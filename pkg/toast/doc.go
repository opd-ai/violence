// Package toast provides a notification queue system for action feedback.
//
// The toast system displays important game events like item pickups, level ups,
// achievements, and quest updates using animated notifications that appear at
// screen edges and auto-dismiss after a configurable duration.
//
// Features:
//   - Priority-based queue (critical notifications shown first)
//   - Stacking with configurable max visible count
//   - Slide-in/fade-out animations with easing
//   - Genre-specific visual styling
//   - Icon support for different notification types
//   - Duration-based auto-dismissal
//
// Example usage:
//
//	// Create toast system
//	sys := toast.NewSystem("fantasy")
//
//	// Queue notifications
//	toast.Queue(world, "item", "Picked up Health Potion", toast.PriorityNormal)
//	toast.Queue(world, "levelup", "Level Up! You are now level 5", toast.PriorityCritical)
//	toast.Queue(world, "achievement", "Achievement Unlocked: First Blood", toast.PriorityHigh)
//
// The system automatically manages queue overflow by discarding oldest low-priority
// notifications when the queue is full.
package toast
