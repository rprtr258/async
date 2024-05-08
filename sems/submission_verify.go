package main

func SubmissionVerifyTest(submittedTest *SubmittedTest) error {
	test := submittedTest.Test

	scores := make([]int, len(test.Questions))
	for i := 0; i < len(test.Questions); i++ {
		question := test.Questions[i]
		submittedQuestion := submittedTest.SubmittedQuestions[i]

		correctAnswers := question.CorrectAnswers
		selectedAnswers := submittedQuestion.SelectedAnswers

		if len(correctAnswers) != len(selectedAnswers) {
			continue
		}

		var nfound int
		for j := 0; j < len(selectedAnswers); j++ {
			selectedAnswer := selectedAnswers[j]

			for k := 0; k < len(correctAnswers); k++ {
				correctAnswer := correctAnswers[k]
				if correctAnswer == selectedAnswer {
					nfound++
					break
				}
			}
		}

		if nfound == len(selectedAnswers) {
			scores[i] = 1
		}
	}
	submittedTest.Scores = scores

	return nil
}

func SubmissionVerifyProgramming(submittedTask *SubmittedProgramming, checkType CheckType) error {
	return nil
}

func SubmissionVerifyStep(step any) {
	switch step := step.(type) {
	case *SubmittedTest:
		SubmissionVerifyTest(step)
	case *SubmittedProgramming:
		step.Error = SubmissionVerifyProgramming(step, CheckTypeTest)
	}
}

func SubmissionVerify(submission *Submission) {
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		SubmissionVerifyStep(submission.SubmittedSteps[i])
	}
}
