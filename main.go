// Mock EC2 Client
// Copyright (C) 2017+ Jonathan Gold and the project contributors
// Written by Jonathan Gold <info@jonathangold.ca> and the project
// contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package ec2mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	InstanceExists               = "InstanceExists"
	AwsErrIncorrectInstanceState = "IncorrectInstanceState"
	AwsErrExceededWaitAttempts   = "exceeded wait attempts"
	InstanceStateNameFilterName  = "instance-state-name"
	TagPrefix                    = "tag:"
	StateChangeDelay             = 5
	DescribeDelay                = 1
	WaiterDelay                  = 5
	MaxWaiterAttempts            = 40
)

type MockEc2Client struct {
	ec2.EC2
	Reservations  []*ec2.Reservation
	InstanceCount int
}

func (c *MockEc2Client) DescribeInstances(in *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	// simulate the time it takes for aws to return
	time.Sleep(DescribeDelay * time.Second)
	var reservations []*ec2.Reservation
	for _, res := range c.Reservations {
		var reservation ec2.Reservation
		for _, inst := range res.Instances {
			instanceID := true
			stateFilter := true
			tagFilter := true
			for _, id := range in.InstanceIds {
				if aws.StringValue(id) != aws.StringValue(inst.InstanceId) {
					instanceID = false
				}
			}
			for _, filter := range in.Filters {
				if aws.StringValue(filter.Name) == InstanceStateNameFilterName {
					stateFilter = false
					for _, val := range filter.Values {
						if aws.StringValue(inst.State.Name) == aws.StringValue(val) {
							stateFilter = true
						}
					}
				}
			}
			for _, tag := range inst.Tags {
				for _, filter := range in.Filters {
					for _, val := range filter.Values {
						if aws.StringValue(filter.Name) == TagPrefix+aws.StringValue(tag.Key) {
							if aws.StringValue(val) != aws.StringValue(tag.Value) {
								tagFilter = false
							}
						}
					}
				}
			}
			if instanceID && stateFilter && tagFilter {
				reservation.Instances = append(reservation.Instances, inst)
			}
		}
		reservations = append(reservations, &reservation)
		break
	}
	return &ec2.DescribeInstancesOutput{
		Reservations: reservations,
	}, nil
}

func (c *MockEc2Client) DescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	time.Sleep(DescribeDelay * time.Second)
	return &ec2.DescribeImagesOutput{
		Images: []*ec2.Image{
			{
				ImageId: aws.String("ami-00000000"),
			},
		},
	}, nil
}

func (c *MockEc2Client) RunInstances(in *ec2.RunInstancesInput) (*ec2.Reservation, error) {
	// TODO: this should technically support running more than one instance
	reservation := &ec2.Reservation{}
	for i := 0; i < int(aws.Int64Value(in.MaxCount)); i++ {
		reservation.Instances = append(reservation.Instances, &ec2.Instance{
			InstanceId: aws.String(string(c.InstanceCount)),
			State: &ec2.InstanceState{
				Name: aws.String(ec2.InstanceStateNamePending),
			},
		})
	}
	// launch a goroutine to change the state after a short delay. We don't
	// have a waitgroup because we need to return right away to match the api.
	c.Reservations = append(c.Reservations, reservation)
	go func() {
		time.Sleep(time.Second * StateChangeDelay)
		for i := 0; i < len(reservation.Instances); i++ {
			reservation.Instances[i].State.Name = aws.String(ec2.InstanceStateNameRunning)
			c.InstanceCount++
		}
	}()
	return reservation, nil
}

func (c *MockEc2Client) CreateTags(in *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error) {
	for _, res := range c.Reservations {
		for _, inst := range res.Instances {
			if aws.StringValue(inst.InstanceId) == aws.StringValue(in.Resources[0]) {
				for _, tag := range in.Tags {
					inst.Tags = append(inst.Tags, tag)
				}
				return &ec2.CreateTagsOutput{}, nil
			}
		}
	}
	return nil, fmt.Errorf("no instance to tag")
}

func (c *MockEc2Client) StartInstances(in *ec2.StartInstancesInput) (*ec2.StartInstancesOutput, error) {
	for _, instanceID := range in.InstanceIds {
		for _, res := range c.Reservations {
			for _, inst := range res.Instances {
				if inst.InstanceId != instanceID {
					continue
				}
				switch aws.StringValue(inst.State.Name) {
				case ec2.InstanceStateNameStopped:
					go func() {
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNamePending)
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameRunning)
					}()
				case ec2.InstanceStateNamePending:
					go func() {
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameRunning)
					}()
				case ec2.InstanceStateNameRunning:
				default:
					return nil, awserr.New(
						AwsErrIncorrectInstanceState,
						"",
						fmt.Errorf(AwsErrIncorrectInstanceState),
					)
				}
				return nil, nil
			}
		}
	}
	return nil, nil
}

func (c *MockEc2Client) StopInstances(in *ec2.StopInstancesInput) (*ec2.StopInstancesOutput, error) {
	for _, instanceID := range in.InstanceIds {
		for _, res := range c.Reservations {
			for _, inst := range res.Instances {
				if inst.InstanceId != instanceID {
					continue
				}
				switch aws.StringValue(inst.State.Name) {
				case ec2.InstanceStateNameRunning:
					go func() {
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameStopping)
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameStopped)
					}()
				case ec2.InstanceStateNameStopping:
					go func() {
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameStopped)
					}()
				case ec2.InstanceStateNameStopped:
				default:
					return nil, awserr.New(AwsErrIncorrectInstanceState, "", nil)
				}
				return nil, nil
			}
		}
	}
	return nil, nil
}

func (c *MockEc2Client) TerminateInstances(in *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	for _, instanceID := range in.InstanceIds {
		for _, res := range c.Reservations {
			for _, inst := range res.Instances {
				if inst.InstanceId != instanceID {
					continue
				}
				switch aws.StringValue(inst.State.Name) {
				case ec2.InstanceStateNameRunning,
					ec2.InstanceStateNameStopped,
					ec2.InstanceStateNameStopping,
					ec2.InstanceStateNamePending:
					go func() {
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameShuttingDown)
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameTerminated)
					}()
				case ec2.InstanceStateNameShuttingDown:
					go func() {
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameTerminated)
					}()
				default:
					return nil, awserr.New(AwsErrIncorrectInstanceState, "", nil)
				}
				return nil, nil
			}
		}
	}
	return nil, nil
}

func (c *MockEc2Client) WaitUntilInstanceRunningWithContext(ctx aws.Context, in *ec2.DescribeInstancesInput, op ...request.WaiterOption) error {
	return c.waiter(ctx, in, ec2.InstanceStateNameRunning)
}

func (c *MockEc2Client) WaitUntilInstanceStoppedWithContext(ctx aws.Context, in *ec2.DescribeInstancesInput, op ...request.WaiterOption) error {
	return c.waiter(ctx, in, ec2.InstanceStateNameStopped)
}

func (c *MockEc2Client) WaitUntilInstanceExistsWithContext(ctx aws.Context, in *ec2.DescribeInstancesInput, op ...request.WaiterOption) error {
	return c.waiter(ctx, in, InstanceExists)
}

func (c *MockEc2Client) waiter(ctx context.Context, in *ec2.DescribeInstancesInput, state string) error {
	errChan := make(chan error)
	defer close(errChan)
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	closeChan := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(closeChan)
		for i := 0; i < MaxWaiterAttempts; i++ {
			time.Sleep(WaiterDelay * time.Second)
			diOutput, err := c.DescribeInstances(in)
			if err != nil {
				select {
				case errChan <- err:
				case <-closeChan:
				}
				return
			}
			for _, res := range diOutput.Reservations {
				for _, inst := range res.Instances {
					if aws.StringValue(inst.State.Name) == state {
						return
					}
				}
			}
			if state == InstanceExists && len(diOutput.Reservations) > 0 {
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-closeChan:
				return
			default:
			}
		}
		select {
		case errChan <- awserr.New(
			request.WaiterResourceNotReadyErrorCode,
			AwsErrExceededWaitAttempts,
			fmt.Errorf("exceded wait attempts"),
		):
			return
		case <-closeChan:
		}
		return

	}()
	select {
	case err, ok := <-errChan:
		if !ok {
			return fmt.Errorf("channel closed unexpectedly")
		}
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return awserr.New(request.CanceledErrorCode, "waiter context cancelled", nil)
	case <-closeChan:
		return nil
	}
	return nil
}
