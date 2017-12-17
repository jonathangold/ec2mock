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
