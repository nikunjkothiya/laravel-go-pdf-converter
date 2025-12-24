package worker

import (
	"context"
	"encoding/json"
	"runtime"
	"sync"
	"time"

	"github.com/nikunjkothiya/gopdfconv/internal/converter"
	"github.com/nikunjkothiya/gopdfconv/internal/pdf"
)

// Job represents a conversion task
type Job struct {
	ID         string
	InputPath  string
	OutputPath string
	Format     converter.FormatType
	Options    pdf.Options
}

// JobResult represents the result of a conversion job
type JobResult struct {
	Job         Job           `json:"job"`
	Success     bool          `json:"success"`
	Error       string        `json:"error,omitempty"`
	ProcessTime time.Duration `json:"process_time_ns"`
	OutputSize  int64         `json:"output_size_bytes"`
}

// Pool manages a pool of workers for concurrent file processing
type Pool struct {
	workers    int
	jobQueue   chan Job
	results    chan JobResult
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	mu               sync.Mutex
	isRunning        bool
	libreOfficePath  string
	native           bool
}

// NewPool creates a new worker pool
func NewPool(workers int, libreOfficePath string) *Pool {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	// Limit to reasonable number
	if workers > 16 {
		workers = 16
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		workers:         workers,
		jobQueue:        make(chan Job, workers*2),
		results:         make(chan JobResult, workers*2),
		ctx:             ctx,
		cancel:          cancel,
		libreOfficePath: libreOfficePath,
	}
}

// Start begins the worker pool
func (p *Pool) Start() {
	p.mu.Lock()
	if p.isRunning {
		p.mu.Unlock()
		return
	}
	p.isRunning = true
	p.mu.Unlock()

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker is the goroutine that processes jobs
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case job, ok := <-p.jobQueue:
			if !ok {
				return
			}
			result := p.processJob(job)
			select {
			case p.results <- result:
			case <-p.ctx.Done():
				return
			}
		}
	}
}

// processJob performs the actual conversion
func (p *Pool) processJob(job Job) JobResult {
	start := time.Now()
	result := JobResult{
		Job: job,
	}

	// Detect format if auto
	format := job.Format
	if format == converter.FormatAuto {
		format = converter.DetectFormat(job.InputPath)
	}

	var err error

	switch format {
	case converter.FormatCSV:
		csvConverter := converter.NewCSVConverter()
		err = csvConverter.Convert(job.InputPath, job.OutputPath, job.Options)

	case converter.FormatXLSX, converter.FormatXLS:
		// For XLSX, try native first. For XLS, try LibreOffice first if available.
		if format == converter.FormatXLS {
			pptxConverter := converter.NewPPTXConverter()
			if p.libreOfficePath != "" {
				pptxConverter.SetLibreOfficePath(p.libreOfficePath)
			}
			if pptxConverter.HasLibreOffice() {
				err = pptxConverter.Convert(job.InputPath, job.OutputPath, job.Options)
			} else {
				excelConverter := converter.NewExcelConverter()
				err = excelConverter.Convert(job.InputPath, job.OutputPath, job.Options)
			}
		} else {
			excelConverter := converter.NewExcelConverter()
			err = excelConverter.Convert(job.InputPath, job.OutputPath, job.Options)
		}

	case converter.FormatPPTX:
		pptxConverter := converter.NewPPTXConverter()
		if p.libreOfficePath != "" {
			pptxConverter.SetLibreOfficePath(p.libreOfficePath)
		}
		if p.native {
			pptxConverter.SetUseLibreOffice(false)
		}
		err = pptxConverter.Convert(job.InputPath, job.OutputPath, job.Options)

	case converter.FormatPPT:
		// Check if LibreOffice is available for better fidelity
		pptxConverter := converter.NewPPTXConverter()
		if p.libreOfficePath != "" {
			pptxConverter.SetLibreOfficePath(p.libreOfficePath)
		}
		if pptxConverter.HasLibreOffice() && !p.native {
			loConverter := converter.NewLibreOfficeConverter(pptxConverter.GetLibreOfficePath())
			err = loConverter.Convert(job.InputPath, job.OutputPath)
		} else {
			// Fall back to native PPT parser (text extraction only)
			pptConverter := converter.NewPPTConverter()
			err = pptConverter.Convert(job.InputPath, job.OutputPath, job.Options)
		}

	default:
		result.Success = false
		result.Error = "Unsupported format: " + string(format)
		return result
	}

	result.ProcessTime = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
	} else {
		result.Success = true
	}

	return result
}

// Submit adds a job to the queue
func (p *Pool) Submit(job Job) {
	select {
	case p.jobQueue <- job:
	case <-p.ctx.Done():
	}
}

// Results returns the results channel
func (p *Pool) Results() <-chan JobResult {
	return p.results
}

// Stop gracefully stops the worker pool
func (p *Pool) Stop() {
	p.mu.Lock()
	if !p.isRunning {
		p.mu.Unlock()
		return
	}
	p.isRunning = false
	p.mu.Unlock()

	close(p.jobQueue)
	p.cancel()
	p.wg.Wait()
	close(p.results)
}

// Wait blocks until all jobs are processed
func (p *Pool) Wait() {
	p.wg.Wait()
}

// BatchConvert performs batch conversion with the worker pool
func BatchConvert(jobs []Job, workers int, libreOfficePath string, native bool) []JobResult {
	pool := NewPool(workers, libreOfficePath)
	pool.native = native
	pool.Start()

	// Submit all jobs
	go func() {
		for _, job := range jobs {
			pool.Submit(job)
		}
		// Close the job queue after all jobs are submitted
		close(pool.jobQueue)
	}()

	// Collect results
	var results []JobResult
	resultCount := 0
	expectedCount := len(jobs)

	for result := range pool.results {
		results = append(results, result)
		resultCount++
		if resultCount >= expectedCount {
			break
		}
	}

	pool.cancel()
	pool.wg.Wait()

	return results
}

// BatchResult summarizes batch conversion results
type BatchResult struct {
	TotalJobs    int           `json:"total_jobs"`
	Successful   int           `json:"successful"`
	Failed       int           `json:"failed"`
	TotalTime    time.Duration `json:"total_time_ns"`
	Results      []JobResult   `json:"results"`
}

// ToJSON returns the batch result as JSON
func (br BatchResult) ToJSON() string {
	data, _ := json.MarshalIndent(br, "", "  ")
	return string(data)
}

// RunBatch executes a batch conversion and returns summarized results
func RunBatch(jobs []Job, workers int, libreOfficePath string, native bool) BatchResult {
	start := time.Now()
	results := BatchConvert(jobs, workers, libreOfficePath, native)

	batch := BatchResult{
		TotalJobs: len(jobs),
		TotalTime: time.Since(start),
		Results:   results,
	}

	for _, r := range results {
		if r.Success {
			batch.Successful++
		} else {
			batch.Failed++
		}
	}

	return batch
}
