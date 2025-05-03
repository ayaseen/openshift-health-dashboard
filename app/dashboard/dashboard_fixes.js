// Fix for CircularProgress component in Dashboard.js
const CircularProgress = ({ value }) => {
    // Make sure value is a number and default to 0 if not
    const percentage = parseFloat(value) || 0;
    const rating = getScoreRating(percentage);
    const color = getScoreColor(percentage);

    // Calculate circle properties
    const radius = 70;
    const stroke = 14;
    const normalizedRadius = radius - stroke / 2;
    const circumference = normalizedRadius * 2 * Math.PI;
    const strokeDashoffset = circumference - (percentage / 100) * circumference;

    return (
        <div className="relative flex items-center justify-center my-8">
            <svg height={radius * 2} width={radius * 2} className="transform -rotate-90">
                <circle
                    stroke="#e5e7eb"
                    fill="transparent"
                    strokeWidth={stroke}
                    r={normalizedRadius}
                    cx={radius}
                    cy={radius}
                />
                <circle
                    stroke={color}
                    fill="transparent"
                    strokeWidth={stroke}
                    strokeDasharray={`${circumference} ${circumference}`}
                    style={{ strokeDashoffset }}
                    strokeLinecap="round"
                    r={normalizedRadius}
                    cx={radius}
                    cy={radius}
                />
            </svg>
            <div className="absolute flex flex-col items-center justify-center">
                <span className="text-3xl font-bold">{Math.round(percentage)}%</span>
                <span className="text-sm text-gray-500">{rating}</span>
            </div>
        </div>
    );
};

// Fix for ProgressBar component in Dashboard.js
const ProgressBar = ({ name, score }) => {
    // Make sure score is a number and default to 0 if not
    const safeScore = parseInt(score, 10) || 0;
    const color = getScoreColor(safeScore);

    return (
        <div className="mb-6">
            <div className="flex justify-between mb-2">
                <span className="text-sm font-medium text-gray-700">{name}</span>
                <span className="text-sm font-medium text-gray-700">{safeScore}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2.5">
                <div
                    className="h-2.5 rounded-full"
                    style={{
                        width: `${safeScore}%`,
                        backgroundColor: color
                    }}
                ></div>
            </div>
        </div>
    );
};
