// Package threat provides a threat indicator system for information hierarchy in the UI.
//
// The threat system addresses UI/UX Problem #4 (Poor Information Hierarchy) by adding
// visual prominence to critical threats. It highlights:
//   - Enemies that are actively attacking the player (pulsing threat border)
//   - Off-screen damage sources (directional threat arrows at screen edges)
//   - Threat level scaling based on proximity and aggression
//
// Integration:
//   - Add ThreatComponent to enemy entities that should display threat indicators
//   - Call ThreatSystem.MarkThreat() when an entity deals damage to the player
//   - Call ThreatSystem.Render() during the Draw phase after world rendering
//
// The system supports all 5 genres with distinct visual styles:
//   - Fantasy: golden/orange warm threat colors
//   - Sci-Fi: cyan/blue electric threat effects
//   - Horror: dark red ominous pulses
//   - Cyberpunk: neon magenta/pink glitch effects
//   - Post-Apocalyptic: rust orange warning indicators
package threat
