export const getScoreColor = (score) => {
    if (score >= 90) return '#10B981'; // Green
    if (score >= 70) return '#22C55E'; // Light Green
    if (score >= 50) return '#F59E0B'; // Amber
    return '#EF4444'; // Red
};

export const getScoreRating = (score) => {
    if (score >= 90) return 'Excellent';
    if (score >= 70) return 'Good';
    if (score >= 50) return 'Fair';
    return 'Poor';
};

export const getHealthAssessment = (score) => {
    if (score >= 80) return 'Your cluster is well configured and follows best practices.';
    if (score >= 60) return 'Your cluster is generally healthy but has some areas needing attention.';
    return 'Your cluster needs significant improvements to meet best practices.';
};